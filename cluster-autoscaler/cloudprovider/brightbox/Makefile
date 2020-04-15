export BUILD_TAGS=brightbox
export REGISTRY=brightbox
ifndef TAG
	override TAG=dev
endif
export TAG

.PHONY: deploy
deploy: examples/config.rb build
	ruby $< | \
	helm template release stable/cluster-autoscaler \
	--namespace kube-system -f - | \
	kubectl -n kube-system apply -f -

.PHONY: remove
remove: examples/config.rb
	ruby $< | \
	helm template release stable/cluster-autoscaler \
	--namespace kube-system -f - | \
	kubectl -n kube-system delete -f -

.PHONY: secret
secret: ${HOME}/.docker/config.json
	-kubectl create secret generic regcred \
	    --from-file=.dockerconfigjson=$? \
	    --type=kubernetes.io/dockerconfigjson

../../cluster-autoscaler: brightbox_cloud_provider.go brightbox_node_group.go brightbox_user_data.go
	$(MAKE) -C $(@D) container
	docker push ${REGISTRY}/cluster-autoscaler-${BUILD_TAGS}:${TAG}

.PHONY: build
build: ../../cluster-autoscaler

.PHONY: clean
clean: 
	$(MAKE) -C ../.. $@
