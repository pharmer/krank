admission_control: '{{ ADMISSION_CONTROL }}'
allocate_node_cidrs: '{{ ALLOCATE_NODE_CIDRS }}'
cluster_cidr: '{{ CLUSTER_IP_RANGE }}'
dns_domain: '{{ DNS_DOMAIN }}'
dns_replicas: '{{ DNS_REPLICAS | integer }}'
dns_server: '{{ DNS_SERVER_IP }}'
elasticsearch_replicas: '{{ ELASTICSEARCH_LOGGING_REPLICAS | integer }}'
enable_cluster_alert: '{{ ENABLE_CLUSTER_ALERT }}'
enable_cluster_dns: '{{ ENABLE_CLUSTER_DNS }}'
enable_cluster_logging: '{{ ENABLE_CLUSTER_LOGGING }}'
enable_cluster_monitoring: '{{ ENABLE_CLUSTER_MONITORING }}'
enable_cluster_registry: 'false'                      {# REMOVE #}
enable_cluster_ui: 'false'                            {# REMOVE #}
enable_third_party_resource: '{{ ENABLE_THIRD_PARTY_RESOURCE }}'
enable_l7_loadbalancing: 'none'                      {# REMOVE #}
enable_manifest_url: '{{ ENABLE_MANIFEST_URL }}'
enable_node_logging: '{{ ENABLE_NODE_LOGGING }}'
hairpin_mode: '{{ HAIRPIN_MODE }}'
instance_prefix: '{{ INSTANCE_PREFIX }}'
kubernetes_master_name: {{ KUBERNETES_MASTER_NAME }}
logging_destination: '{{ LOGGING_DESTINATION }}'
manifest_url: '{{ MANIFEST_URL }}'
manifest_url_header: '{{ MANIFEST_URL_HEADER }}'
master_internal_ip : {{ MASTER_INTERNAL_IP }}
network_provider: '{{ NETWORK_PROVIDER }}'
node_instance_prefix: '{{ INSTANCE_PREFIX }}-node'
node_tags: '{{ INSTANCE_PREFIX }}-node'
num_nodes: {{ NUM_NODES | integer }}
runtime_config: '{{ RUNTIME_CONFIG }}'
service_cluster_ip_range: '{{ SERVICE_CLUSTER_IP_RANGE }}'
zone: {{ ZONE }}
{% if ENABLE_CLUSTER_SECURITY %}enable_cluster_security: '{{ENABLE_CLUSTER_SECURITY}}'{% endif %}
{% if KUBELET_PORT %}kubelet_port: '{{KUBELET_PORT}}'{% endif %}
{% if TERMINATED_POD_GC_THRESHOLD %}terminated_pod_gc_threshold: '{{TERMINATED_POD_GC_THRESHOLD}}'{% endif %}
{% if ENABLE_CUSTOM_METRICS %}enable_custom_metrics: '{{ENABLE_CUSTOM_METRICS}}'{% endif %}
{% if ENABLE_CLUSTER_VPN %}enable_cluster_vpn: '{{ENABLE_CLUSTER_VPN}}'{% endif %}
{% if VPN_PSK %}vpn_psk: '{{VPN_PSK}}'{% endif %}
{% if APPSCODE_API_GRPC_ENDPOINT %}appscode_api_grpc_endpoint: '{{APPSCODE_API_GRPC_ENDPOINT}}'{% endif %}
{% if APPSCODE_API_HTTP_ENDPOINT %}appscode_api_http_endpoint: '{{APPSCODE_API_HTTP_ENDPOINT}}'{% endif %}
{% if APPSCODE_CLUSTER_ROOT_DOMAIN %}appscode_cluster_root_domain: '{{APPSCODE_CLUSTER_ROOT_DOMAIN}}'{% endif %}
{% if APPSCODE_NS %}appscode_ns: '{{APPSCODE_NS}}'{% endif %}
{% if KUBE_UID %}kube_uid: '{{KUBE_UID}}'{% endif %}
{% if NODE_LABELS %}node_labels: '{{NODE_LABELS}}'{% endif %}
{% if ENABLE_NODE_PROBLEM_DETECTOR %}enable_node_problem_detector: '{{ENABLE_NODE_PROBLEM_DETECTOR}}'{% endif %}
{% if NETWORK_POLICY_PROVIDER %}network_policy_provider: '{{NETWORK_POLICY_PROVIDER}}'{% endif %}
{% autoescape off %}
{% if EVICTION_HARD %}eviction_hard: '{{EVICTION_HARD}}'{% endif %}
{% endautoescape %}
federations_domain_map: ''
non_masquerade_cidr: ''  # aws only
{% if ENABLE_RESCHEDULER %}enable_rescheduler: '{{ENABLE_RESCHEDULER}}'{% endif %}
{% if INITIAL_ETCD_CLUSTER %}initial_etcd_cluster: '{{INITIAL_ETCD_CLUSTER}}'{% endif %}
{% if HOSTNAME %}hostname: '{{HOSTNAME}}'{% endif %}
{% if HOST_IFACE %}host_iface: '{{HOST_IFACE}}'{% endif %}
{% if ENABLE_RBAC_AUTHZ %}enable_rbac_authz: '{{ ENABLE_RBAC_AUTHZ }}'{% endif %}
{% if APPSCODE_CLUSTER_USER %}appscode_cluster_user: '{{ APPSCODE_CLUSTER_USER }}'{% endif %}
{% if APPSCODE_CLUSTER_CREATOR %}appscode_cluster_creator: '{{ APPSCODE_CLUSTER_CREATOR }}'{% endif %}
{% if HOSTFACTS_AUTH_TOKEN %}hostfacts_auth_token: '{{ HOSTFACTS_AUTH_TOKEN }}'{% endif %}
{% if ENABLE_APISERVER_BASIC_AUDIT %}enable_apiserver_basic_audit: '{{ ENABLE_APISERVER_BASIC_AUDIT }}'{% endif %}
{% if SOFTLOCKUP_PANIC %}softlockup_panic: '{{ SOFTLOCKUP_PANIC }}'{% endif %}
