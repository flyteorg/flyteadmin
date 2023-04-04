package testutils

import "github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/admin"

var ExecutionQueueAttributes = &admin.MatchingAttributes{
	Target: &admin.MatchingAttributes_ExecutionQueueAttributes{
		ExecutionQueueAttributes: &admin.ExecutionQueueAttributes{
			Tags: []string{
				"foo", "bar", "baz",
			},
		},
	},
}

var WorkflowExecutionConfigSample = &admin.MatchingAttributes{
	Target: &admin.MatchingAttributes_WorkflowExecutionConfig{
		WorkflowExecutionConfig: &admin.WorkflowExecutionConfig{
			MaxParallelism: 5,
			RawOutputDataConfig: &admin.RawOutputDataConfig{
				OutputLocationPrefix: "s3://test-bucket",
			},
			Labels: &admin.Labels{
				Values: map[string]string{"lab1": "val1"},
			},
			Annotations: nil,
		},
	},
}

var TaskResourcesSample = &admin.MatchingAttributes{
	Target: &admin.MatchingAttributes_TaskResourceAttributes{
		TaskResourceAttributes: &admin.TaskResourceAttributes{
			Defaults: &admin.TaskResourceSpec{
				Cpu:              "1",
				Gpu:              "2",
				Memory:           "100Mi",
				EphemeralStorage: "100Gi",
			},
			Limits: &admin.TaskResourceSpec{
				Cpu:              "2",
				Gpu:              "3",
				Memory:           "200Mi",
				EphemeralStorage: "300Gi",
			},
		},
	},
}
