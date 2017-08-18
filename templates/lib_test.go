package templates

import (
	"fmt"
	"testing"

	"github.com/appscode/go/flags"
	"github.com/flosch/pongo2"
	"github.com/stretchr/testify/assert"
)

func TestContext(t *testing.T) {
	type AnotherType struct {
		Name  string
		Value string
	}

	type AType struct {
		Name        string
		Value       int
		AnotherType *AnotherType
	}

	a := &AType{
		Name:  "one-name",
		Value: 5,
		AnotherType: &AnotherType{
			Name:  "another inner type",
			Value: "99",
		},
	}

	ctx, err := Context(a)
	assert.Nil(t, err)

	c := &pongo2.Context{
		"Name":  "one-name",
		"Value": 5,
		"AnotherType": map[string]interface{}{
			"Name":  "another inner type",
			"Value": "99",
		},
	}
	assert.Equal(t, ctx, c)
}

type Options struct {
	Sticky         bool
	SSLCert        bool
	OverridePath   bool
	DefaultBackend *Backend
	HttpsService   []*Service
	HttpService    []*Service
}

type Backend struct {
	Name      string
	Endpoints []*Endpoint
}

type Endpoint struct {
	Name string
	IP   string
	Port string
}

type Service struct {
	Name     string
	AclMatch string
	Host     string
	Backends *Backend
}

func TestHAProxyTemplate(t *testing.T) {
	flags.SetLogLevel(10)

	ep := []*Endpoint{
		{
			Name: "aa",
			IP:   "1.1.1.1",
			Port: "8080",
		},
		{
			Name: "bb",
			IP:   "1.1.1.2",
			Port: "9090",
		},
	}
	httpService := []*Service{
		{
			Name: "one",
			Host: "a.b.com",
			Backends: &Backend{
				Name:      "server-1",
				Endpoints: ep,
			},
		},
		{
			Name: "two",
			Backends: &Backend{
				Name:      "server-1",
				Endpoints: ep,
			},
		},
		{
			Name:     "one",
			AclMatch: "/beg",
			Host:     "a.b.com",
			Backends: &Backend{
				Name:      "server-1",
				Endpoints: ep,
			},
		},
	}

	op := &Options{
		Sticky: true,
		//SSLCert: true,
		OverridePath: true,

		DefaultBackend: &Backend{
			Name: "bk-one",
			Endpoints: []*Endpoint{
				{
					Name: "aa",
					IP:   "1.1.1.1",
					Port: "8080",
				},
				{
					Name: "bb",
					IP:   "1.1.1.2",
					Port: "9090",
				},
			},
		},

		HttpService:  httpService,
		HttpsService: httpService,
	}

	c, err := Context(op)
	assert.Nil(t, err)

	s, err := Render(c, "lb/ingress-temp.cfg")
	fmt.Println(err)

	fmt.Println(s)
}

func TestMap(t *testing.T) {
	flags.SetLogLevel(5)
	var temp string = `
{% for k, v in mp %}
{{k}} {{v}}
{% endfor %}`

	mp := map[string]string{
		"hello": "world",
		"world": "hello",
	}

	ctx := &pongo2.Context{
		"mp": mp,
	}

	fmt.Println(RenderString(ctx, temp))
}
