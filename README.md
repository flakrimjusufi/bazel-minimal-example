## A minimal project with Bazel

### Purpose

- Following the opened issue in https://github.com/bazelbuild/bazel-gazelle/issues/1289
- To demonstrate that Bazel is failing when google-related dependencies are upgraded

### How to reproduce the error

Run the following command:
~~~
    make bazel-test 
~~~

This command with run the test cases which contains the tag="fast" while executing the following: 
~~~
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
~~~

#### To run the project with Bazel

~~~
    make bazel-run 
~~~

#### To clean project files that were built using with Bazel

~~~
    make bazel-clean 
~~~


#### To set up the project with Bazel 

~~~
    make bazel-setup
~~~

#### To update dependencies with Bazel

~~~
    make bazel-update-deps 
~~~


