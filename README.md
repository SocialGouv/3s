# 3S (StatefulScalerService)

3S (StatefulScalerService) is a Kubernetes operator designed to dynamically adjust service selectors in response to scaling events of a StatefulSet. This ensures that services always correctly target the appropriate pods, enabling efficient load balancing and seamless scaling. The primary purpose of 3S is to optimize consistent hashing distribution over scaling events, such as minimizing cache loss in distributed services like BuildKit.

## Installation

3S can be easily deployed using the provided Helm chart:

```shell
git clone https://github.com/SocialGouv/3s.git
cd 3s/charts/3s
helm upgrade --install 3s . --namespace your-namespace \
  --set podSelector="app.kubernetes.io/name=buildkit-service" \
  --set statefulsetNameFormat="buildkit-service-%d" \
  --set servicePrefix="svc-"
```

## Features

* **Dynamic Service Selector Updates:** Automatically updates service selectors based on StatefulSet pod scaling events to maintain efficient load balancing.
* **Optimization for Consistent Hashing:** Designed to minimize disruption in consistent hashing setups, reducing cache invalidation and improving the efficiency of distributed services.
* **High Availability:** Ensures services are always directed to available pods, enhancing the reliability of your deployments.
* **Customizable:** Easy configuration through environment variables to match your specific Kubernetes environment needs.
* **Lightweight Design:** 3S is built to be minimal and efficient, ensuring minimal overhead on your Kubernetes cluster.
* **Helm Chart Deployment:** Easily deployable via an accompanying Helm chart for quick and easy installation.

## Prerequisites

* Kubernetes cluster
* HPA on a StatefulSet

## Configuration

Configure 3S using the following environment variables:

* `SSS_POD_SELECTOR`: The label selector for the target StatefulSet pods. eg: `app.kubernetes.io/name=buildkit-service`
* `SSS_SERVICE_PREFIX`: The service name prefix. eg: `svc-`
* `SSS_SERVICE_SELECTOR`: The label selector for the services to be adjusted. Optional, the filter will be on SSS_SERVICE_PREFIX anyway.
* `SSS_NAMESPACE`: The namespace where the operator and target StatefulSet reside. Optional, default to where the operator is deployed.

## Usage

Once deployed, 3S automatically monitors for scaling events of the specified StatefulSet and updates the selectors of the designated services. This ensures that traffic distribution adjusts dynamically, optimizing the consistent hashing distribution and minimizing potential cache loss during scaling operations.

Note: Each service you wish to manage with 3S requires its own operator deployment, allowing for fine-grained control over the scaling and service selector adjustment process.

## Contributing

We welcome contributions! Please feel free to submit pull requests or open issues on our GitHub repository.

## License

This project is licensed under the MIT License - see the LICENSE file for details.