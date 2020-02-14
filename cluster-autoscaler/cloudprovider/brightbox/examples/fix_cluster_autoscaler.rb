require 'yaml'

result = []
YAML.load_stream(ARGF) do |resource|
  case resource['kind']
  when 'Service'
    next
  when 'ClusterRole'
    resource['rules'] <<
      {
        'apiGroups' => ['coordination.k8s.io'],
        'resources' => ['leases'],
        'verbs' => ['create']
      } <<
      {
        'apiGroups' => ['coordination.k8s.io'],
        'resourceNames' => ['cluster-autoscaler'],
        'resources' => ['leases'],
        'verbs' => %w[get update]
      }
  end
  result << resource
end
print YAML.dump_stream(*result)
