// SPDX-FileCopyrightText: 2022 Red Hat, Inc. <sd-mt-sre@redhat.com>
//
// SPDX-License-Identifier: Apache-2.0

package ocm

import (
	"context"
	"fmt"
	"strings"

	"github.com/apex/log"
	"github.com/apex/log/handlers/discard"
	sdk "github.com/openshift-online/ocm-sdk-go"
	cmv1 "github.com/openshift-online/ocm-sdk-go/clustersmgmt/v1"
)

func NewAddon(addon *cmv1.AddOn, opts ...AddonOption) Addon {
	ao := Addon{
		addon: addon,
	}

	ao.cfg.Option(opts...)
	ao.cfg.Default()

	return ao
}

// Addon wraps an 'ocm-sdk-go' AddOn object.
type Addon struct {
	addon   *cmv1.AddOn
	cfg     AddonConfig
	version *AddonVersion
}

func (a *Addon) ID() string   { return a.addon.ID() }
func (a *Addon) Name() string { return a.addon.Name() }

func (a *Addon) ProvideRowData() map[string]interface{} {
	result := map[string]interface{}{
		"Description":            a.addon.Description(),
		"Docs Link":              a.addon.DocsLink(),
		"Enabled":                a.addon.Enabled(),
		"Has External Resources": a.addon.HasExternalResources(),
		"Hidden":                 a.addon.Hidden(),
		"Icon":                   a.addon.Icon(),
		"ID":                     a.ID(),
		"Install Mode":           a.addon.InstallMode(),
		"Label":                  a.addon.Label(),
		"Name":                   a.Name(),
		"Operator Name":          a.addon.OperatorName(),
		"Resource Cost":          a.addon.ResourceCost(),
		"Resource Name":          a.addon.ResourceName(),
		"Version ID":             a.addon.Version().ID(),
	}

	for k, v := range a.version.ProvideRowData() {
		result["Version "+k] = v
	}

	return result
}

func (a *Addon) WithVersion(ctx context.Context) (*Addon, error) {
	version := a.addon.Version().ID()

	trace := a.cfg.Logger.
		WithFields(log.Fields{
			"addon":   a.ID(),
			"version": version,
		}).
		Trace("requesting version information")
	defer trace.Stop(nil)

	ver, err := a.cfg.Conn.
		ClustersMgmt().
		V1().
		Addons().
		Addon(a.ID()).
		Versions().
		Version(version).
		Get().
		SendContext(ctx)
	if err != nil {
		return a, fmt.Errorf("requesting addon: %w", err)
	}

	a.version = &AddonVersion{
		ver: ver.Body(),
	}

	return a, nil
}

type AddonConfig struct {
	Conn   *sdk.Connection
	Logger log.Interface
}

func (c *AddonConfig) Option(opts ...AddonOption) {
	for _, opt := range opts {
		opt.ConfigureAddon(c)
	}
}

func (c *AddonConfig) Default() {
	if c.Logger == nil {
		c.Logger = &log.Logger{
			Handler: discard.New(),
		}
	}
}

type AddonOption interface {
	ConfigureAddon(*AddonConfig)
}

type AddonVersion struct {
	ver *cmv1.AddOnVersion
}

func (v *AddonVersion) ProvideRowData() map[string]interface{} {
	if v == nil {
		return map[string]interface{}{}
	}

	return map[string]interface{}{
		"Available Upgrades": strings.Join(v.ver.AvailableUpgrades(), ", "),
		"Channel":            v.ver.Channel(),
		"Enabled":            v.ver.Enabled(),
		"Source Image":       v.ver.SourceImage(),
	}
}

type AddonParameter struct {
	param *cmv1.AddOnParameter
}

func (ap *AddonParameter) ProvideRowData() map[string]interface{} {
	if ap == nil {
		return map[string]interface{}{}
	}

	return map[string]interface{}{
		"Default Value":            ap.param.DefaultValue(),
		"Description":              ap.param.Description(),
		"Editable":                 ap.param.Editable(),
		"Enabled":                  ap.param.Enabled(),
		"ID":                       ap.param.ID(),
		"Name":                     ap.param.Name(),
		"Options":                  strings.Join(ap.options(), ", "),
		"Required":                 ap.param.Required(),
		"Validation":               ap.param.Validation(),
		"Validation Error Message": ap.param.ValidationErrMsg(),
		"Value Type":               ap.param.ValueType(),
	}
}

func (ap *AddonParameter) options() []string {
	options := ap.param.Options()

	result := make([]string, 0, len(options))

	for _, opt := range options {
		result = append(result, fmt.Sprintf("%s: %s", opt.Name(), opt.Value()))
	}

	return result
}

type AddonRequirement struct {
	req *cmv1.AddOnRequirement
}

func (ar *AddonRequirement) ProvideRowData() map[string]interface{} {
	if ar == nil {
		return map[string]interface{}{}
	}

	return map[string]interface{}{
		"Enabled":               ar.req.Enabled(),
		"ID":                    ar.req.ID(),
		"Resource":              ar.req.Resource(),
		"Status Error Messages": strings.Join(ar.req.Status().ErrorMsgs(), ", "),
		"Status Fulfilled":      ar.req.Status().Fulfilled(),
	}
}

type AddonSubOperator struct {
	sub *cmv1.AddOnSubOperator
}

func (aso *AddonSubOperator) ProvideRowData() map[string]interface{} {
	if aso == nil {
		return map[string]interface{}{}
	}

	return map[string]interface{}{
		"Operator Name":      aso.sub.OperatorName(),
		"Operator Namespace": aso.sub.OperatorNamespace(),
		"Enabled":            aso.sub.Enabled(),
	}
}

type CredentialRequest struct {
	req *cmv1.CredentialRequest
}

func (cr *CredentialRequest) ProvideRowData() map[string]interface{} {
	if cr == nil {
		return map[string]interface{}{}
	}

	return map[string]interface{}{
		"Name":               cr.req.Name(),
		"Namespace":          cr.req.Namespace(),
		"Policy Permissions": strings.Join(cr.req.PolicyPermissions(), ", "),
		"Service Account":    cr.req.ServiceAccount(),
	}
}
