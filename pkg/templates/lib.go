package templates

import (
	"encoding/json"
	"os"

	"github.com/appscode/errors"
	"github.com/flosch/pongo2"
)

func AssetBytes(path string) ([]byte, error) {
	bytes, err := Asset(path)
	if err != nil {
		return bytes, errors.FromErr(err).Err()
	}
	return bytes, nil
}

func AssetText(path string) (string, error) {
	s, err := AssetBytes(path)
	if err != nil {
		return "", errors.FromErr(err).Err()
	}
	return string(s), nil
}

func Load(path string) (*pongo2.Template, error) {
	bytes, err := Asset(path)
	if err != nil {
		return nil, errors.FromErr(err).Err()
	}
	tpl, err := pongo2.FromString(string(bytes))
	if err != nil {
		return nil, errors.FromErr(err).Err()
	}
	return tpl, nil
}

func Render(ctx *pongo2.Context, in string) (string, error) {
	tpl, err := Load(in)
	if err != nil {
		return "", errors.FromErr(err).Err()
	}
	out, err := tpl.Execute(*ctx)
	if err != nil {
		return "", errors.FromErr(err).Err()
	}
	return out, nil
}

func RenderString(ctx *pongo2.Context, temp string) (string, error) {
	tpl, err := pongo2.FromString(temp)
	if err != nil {
		return "", errors.FromErr(err).Err()
	}
	out, err := tpl.Execute(*ctx)
	if err != nil {
		return "", errors.FromErr(err).Err()
	}
	return out, nil
}

func Write(out string, ctx *pongo2.Context, in string) error {
	f, err := os.Create(out)
	if err != nil {
		return errors.FromErr(err).Err()
	}
	defer f.Close()

	tpl, _ := Load(in)
	err = tpl.ExecuteWriter(*ctx, f)
	if err != nil {
		return errors.FromErr(err).Err()
	}
	return nil
}

func Context(s interface{}) (*pongo2.Context, error) {
	d, err := json.Marshal(s)
	if err != nil {
		return nil, err
	}
	ctx := &pongo2.Context{}
	err = json.Unmarshal(d, ctx)
	if err != nil {
		return nil, errors.FromErr(err).Err()
	}
	return ctx, nil
}
