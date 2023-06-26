## Zarf Package for Big Bang

### Description

This folder and associated Zarf package contains DUBBD (Defense Unicorns BigBang Distribution), which
comprises of 4 Zarf components: `load-certs`, `preflight`, `download-flux`, and `bigbang`. See the DUBBD
repo for further context and details: [uds-package-dubbd](https://github.com/defenseunicorns/uds-package-dubbd).
Via DUBBD, we are able to overlay our own values.yaml on top of DUBBD's overrides, resulting in an _opinionated_
release of [Big Bang](https://docs-bigbang.dso.mil/latest/), which includes only certain Big Bang applications.

### Additional Customization

Big Bang defines cascading Helm chart
[values](https://docs-bigbang.dso.mil/latest/docs/understanding-bigbang/configuration/base-config/#Values).
We selectively override/merge our own custom values provided in our
(values.yaml file)[./kustomizations/bigbang/values.yaml].

### Gateway Configuration

Big Bang provides [Istio](https://istio.io/) which we use for a Service Mesh
and also ingress gateways. Big Bang deployments typically define a single
ingress gateway `admin-ingressgateway`. We add two additional ingress
gateways: `dataplane-ingressgateway` and `passthrough-ingressgateway`. The
data plane gateway is used to provide some level of isolation between the Big
Bang services and those added later.

Both the admin-ingressgateway and dataplane-ingressgateway will do TLS
termination at the gateway. The passthrough gateway is used for traffic to
the keycloak service because keycloak insists on doing its own TLS termination.

> ⚠️ **The default k3d load balancer is
> unable to cope with multiple ingress gateways.** We recommend starting k3d
> with `--k3s-arg --disable=servicelb@server:*` and then running Metal LB on
> the cluster. This will allow you to run this package using all the gateways
> specified. For more details see
> [this](https://github.com/keunlee/k3d-metallb-starter-kit). This
> technique is used for the automated
> [unit test](../test/dco_core_package_test.go).

### Dependencies:

At deployment time, this package requires a zarf-config.yaml file present, either in the working directory
or by setting `ZARF_CONFIG` to the location of the zarf-config.yaml file on the filesystem. This config file
specifies the domain and correlating cert/key files for the domain. In short, the zarf-config.yaml must 
contain the following, where key_file and cert_file are the names of local files containing the cert and key:
```bash
package:
  deploy:
    set:
      domain: <domain name, ex: 'vp.bigbang.dev'>
      key_file: <key filename, ex: 'vp.bigbang.dev.key'>
      cert_file: <cert filename, ex: 'vp.bigbang.dev.cert'>
```
