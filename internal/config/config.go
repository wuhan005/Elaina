// Copyright 2024 E99p1ant. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package config

import (
	"github.com/kelseyhightower/envconfig"
	"github.com/pkg/errors"
)

var App struct {
	Password    string `envconfig:"APP_PASSWORD"`
	RuntimeMode string `envconfig:"RUNTIME_MODE" default:"docker"`

	KubernetesServiceHost string `envconfig:"KUBERNETES_SERVICE_HOST"`
	KubernetesNamespace   string `envconfig:"KUBERNETES_NAMESPACE"`
	KubernetesCAData      string `envconfig:"KUBERNETES_CA_DATA"`
	KubernetesCertData    string `envconfig:"KUBERNETES_CERT_DATA"`
	KubernetesKeyData     string `envconfig:"KUBERNETES_KEY_DATA"`
	KubernetesBearerToken string `envconfig:"KUBERNETES_BEARER_TOKEN"`
}

var Postgres struct {
	DSN string `envconfig:"POSTGRES_DSN"`
}

func Init() error {
	if err := envconfig.Process("", &App); err != nil {
		return errors.Wrap(err, "parse app")
	}
	if err := envconfig.Process("", &Postgres); err != nil {
		return errors.Wrap(err, "parse postgres")
	}

	return nil
}
