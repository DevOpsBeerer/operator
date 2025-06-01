# Generate code
generate:
	controller-gen object paths="./api/..."

# Generate CRD manifests
manifests:
	controller-gen crd paths="./..." output:crd:artifacts:config=config/crd/bases
