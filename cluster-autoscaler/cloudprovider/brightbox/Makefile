.PHONY: deploy
deploy: examples/config.rb build
	ruby $< | \
	helm template release stable/cluster-autoscaler \
	--namespace kube-system -f - | \
	kubectl -n kube-system apply -f -

export BUILD_TAGS=brightbox
export REGISTRY=brightbox
export TAG=dev
../../cluster-autoscaler: brightbox_cloud_provider.go brightbox_node_group.go brightbox_user_data.go
	$(MAKE) -C $(@D) container
	docker push ${REGISTRY}/cluster-autoscaler-${BUILD_TAGS}:${TAG}

.PHONY: build
build: ../../cluster-autoscaler

.PHONY: clean
clean: 
	$(MAKE) -C ../.. $@
