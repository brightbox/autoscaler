.PHONY: deploy
deploy: examples/config.rb
	ruby $< | \
	helm template release stable/cluster-autoscaler \
	--namespace kube-system -f - | \
	ruby examples/fix_cluster_autoscaler.rb | \
	kubectl -n kube-system apply -f -
