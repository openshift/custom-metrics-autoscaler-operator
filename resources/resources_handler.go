package resources

import (
	"bytes"
	"embed"

	mf "github.com/manifestival/manifestival"
)

const resourcesPath = "keda.yaml"
const olmResourcesPath = "keda-olm-operator.yaml"
const httpAddonResourcesPath = "keda-http-addon.yaml"
const LastConfigID = "olm-operator.keda.sh/last-applied-configuration"

func GetResourcesManifest() (mf.Manifest, error) {
	kedamf, err := manifestFromEmbed(resourcesPath)
	if err != nil {
		return kedamf, err
	}
	operatormf, err := manifestFromEmbed(olmResourcesPath)
	return kedamf.Append(operatormf), err
}

func GetHTTPAddonResourcesManifest() (mf.Manifest, error) {
	_, path, _, _ := runtime.Caller(0)
	fullPath := filepath.Join(filepath.Dir(path), httpAddonResourcesPath)
	return mf.NewManifest(fullPath, mf.UseLastAppliedConfigAnnotation(LastConfigID))
}
