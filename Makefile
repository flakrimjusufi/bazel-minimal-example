CLUSTER=$1
NAMESPACE=$2


### BAZEL COMMANDS
bazel-clean: ### clean bazel cached files
	@echo "::bazel-clean";
	@bazelisk clean;
.PHONY: bazel-clean

bazel-setup: ### creates the setup for bazel
	@echo "::bazel-setup";
	@bazelisk run --platforms=@io_bazel_rules_go//go/toolchain:linux_amd64 //:gazelle
.PHONY: bazel-setup

bazel-update-deps: ### updates the dependencies if a dependency has changed/added
	@echo "::bazel-update-deps";
	@bazelisk run --platforms=@io_bazel_rules_go//go/toolchain:linux_amd64 //:gazelle -- update;
	@bazelisk run --platforms=@io_bazel_rules_go//go/toolchain:linux_amd64 //:gazelle -- update-repos -from_file=go.mod;
.PHONY: bazel-update-deps

bazel-run: ### runs the project with bazel
	@echo "::bazel-run"; \
    bazelisk run //:gazelle
.PHONY: bazel-run

bazel-test: ### runs the test cases with bazel
	@echo "::bazel-test"; \
    bazelisk test \
        --platform_suffix="bazel-test" \
        --@io_bazel_rules_go//go/config:race \
        --action_env=TESTING=test \
        --define cluster=$CLUSTER \
        --define namespace=$NAMESPACE \
        --test_tag_filters=fast \
        --build_tag_filters=fast \
        --test_output=errors \
        --nocache_test_results \
        //...
.PHONY: bazel-test