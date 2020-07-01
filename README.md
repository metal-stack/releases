# Releases

This repository contains release notes and release image vectors for metal-stack.

**Please use tagged releases of this repository for ensuring stable updates!**

The metal-stack is a compound of microservices, which all have their individual versions and development progresses. Hence, there is a unique image vector for each metal-stack release describing the set of the latest stable microservices that are compatible with each other.

## Release notes for v0.1.0

### firewall-controller v0.1.5

The [firewall-controller](https://github.com/metal-stack/firewall-controller) is a new component that deprecates the [firewall-policy-controller](https://github.com/metal-stack/firewall-policy-controller). It is a kubernetes controller that reconciles objects in the `firewall` namespace of shoot clusters.

- a `firewall` CRD and object is introduced to shoot clusters with statistics about nftables rules and traffic counters for internal/external networks:

    ```bash
    kubectl describe firewalls -n firewall firewall
    ```

- a CRD `ClusterwideNetworkPolicy` is introduced to shoot clusters for policies that target the whole cluster and that are enforced at the firewall; those network policies do not interfere with Calico because and standard kubernetes `NetworkPolicies` without PodSelector are migrated by the controller to new `ClusterwideNetworkPolicies`.

    ```bash
    kubectl get clusterwidenetworkpolicy -A
    ```

- exporters of firewalls are exposed as services into shoot clusters:

    ```bash
    kubectl get svc -n firewall
    ```

- preparation for IDS support with suricata:

    ```bash
    kubectl describe firewalls -n firewall firewall # shows first generic suricata statistics
    ```

- traffic to specific networks can be rate limited from the firewall object
- included with firewall images >= 20200701
- for more information see https://github.com/metal-stack/firewall-controller

### metal-api v0.7.8

- fix for allocate and release ip addresses

### metalctl v0.7.8

- fix led state
- switch contexts and short context
- show firewall age

### gepm v0.11.5

- support for new firewall-controller
- rate limits are adjustable from the shoot spec:

    ```
    kind: InfrastructureConfig
    firewall:
      rateLimits:
        internet: 100 # maximum rate in MBit/s
    ```

- refactoring of cloud profile
