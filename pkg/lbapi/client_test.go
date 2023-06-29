package lbapi

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/shurcooL/graphql"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetLoadBalancer(t *testing.T) {
	respJSON := `{
	"data": {
		"loadBalancer": {
			"id": "loadbal-randovalue",
			"name": "some lb",
			"ports": {
				"edges": [
					{
						"node": {
							"name": "porty",
							"id": "loadprt-randovalue",
							"number": 80,
							"nodeID": "loadbal-randovalue",
							"IPAddresses": [
								{
									"id": "ipamipa-randovalue",
									"ip": "192.168.1.42",
									"reserved": false
								}
							]
						}
					}
				]
			}
		}
	}
}`

	cli := Client{
		gqlCli: mustNewGQLTestClient(respJSON),
	}

	t.Run("bad prefix", func(t *testing.T) {
		lb, err := cli.GetLoadBalancer(context.Background(), "badprefix-test")
		require.Error(t, err)
		require.Nil(t, lb)
		assert.ErrorContains(t, err, "invalid id")
	})

	t.Run("successful query", func(t *testing.T) {
		lb, err := cli.GetLoadBalancer(context.Background(), "loadbal-randovalue")
		require.NoError(t, err)
		require.NotNil(t, lb)

		assert.Equal(t, "loadbal-randovalue", lb.LoadBalancer.ID)
		assert.Equal(t, "some lb", lb.LoadBalancer.Name)

		require.Len(t, lb.LoadBalancer.Ports.Edges, 1)
		assert.Equal(t, "loadprt-randovalue", lb.LoadBalancer.Ports.Edges[0].Node.ID)
		assert.Equal(t, "porty", lb.LoadBalancer.Ports.Edges[0].Node.Name)
		assert.Equal(t, int64(80), lb.LoadBalancer.Ports.Edges[0].Node.Number)
		assert.Empty(t, lb.LoadBalancer.Ports.Edges[0].Node.Pools)

		require.Len(t, lb.LoadBalancer.Ports.Edges[0].Node.IPAddresses, 1)
		assert.Equal(t, "ipamipa-randovalue", lb.LoadBalancer.Ports.Edges[0].Node.IPAddresses[0].ID)
		assert.Equal(t, "192.168.1.42", lb.LoadBalancer.Ports.Edges[0].Node.IPAddresses[0].IP)
		assert.False(t, lb.LoadBalancer.Ports.Edges[0].Node.IPAddresses[0].Reserved)
	})
}

func mustNewGQLTestClient(respJSON string) *graphql.Client {
	mux := http.NewServeMux()
	mux.HandleFunc("/query", func(w http.ResponseWriter, req *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, err := io.WriteString(w, respJSON)
		if err != nil {
			panic(err)
		}
	})

	return graphql.NewClient("/query", &http.Client{Transport: localRoundTripper{handler: mux}})
}

type localRoundTripper struct {
	handler http.Handler
}

func (l localRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	w := httptest.NewRecorder()
	l.handler.ServeHTTP(w, req)

	return w.Result(), nil
}
