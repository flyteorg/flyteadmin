package notifications

import (
	"fmt"

	"strings"

	runtimeInterfaces "github.com/lyft/flyteadmin/pkg/runtime/interfaces"
	"github.com/lyft/flyteidl/gen/pb-go/flyteidl/admin"
)

const executionError = " The execution failed with error: [%s]."

const substitutionParam = "{{ %s }}"
const project = "project"
const domain = "domain"
const name = "name"
const phase = "phase"
const errorPlaceholder = "error"
const workflowProject = "workflow.project"
const workflowDomain = "workflow.domain"
const workflowName = "workflow.name"
const workflowVersion = "workflow.version"
const launchPlanProject = "launch_plan.project"
const launchPlanDomain = "launch_plan.domain"
const launchPlanName = "launch_plan.name"
const launchPlanVersion = "launch_plan.version"
const replaceAllInstances = -1

func substituteEmailParameters(message string, request admin.WorkflowExecutionEventRequest, execution *admin.Execution) string {
	response := strings.Replace(message, fmt.Sprintf(substitutionParam, project), execution.Id.Project, replaceAllInstances)
	response = strings.Replace(response, fmt.Sprintf(substitutionParam, domain), execution.Id.Domain, replaceAllInstances)
	response = strings.Replace(response, fmt.Sprintf(substitutionParam, name), execution.Id.Name, replaceAllInstances)
	response = strings.Replace(response, fmt.Sprintf(substitutionParam, phase),
		strings.ToLower(request.Event.Phase.String()), replaceAllInstances)
	response = strings.Replace(response, fmt.Sprintf(substitutionParam, workflowProject), execution.Closure.WorkflowId.Project, replaceAllInstances)
	response = strings.Replace(response, fmt.Sprintf(substitutionParam, workflowDomain), execution.Closure.WorkflowId.Domain, replaceAllInstances)
	response = strings.Replace(response, fmt.Sprintf(substitutionParam, workflowName), execution.Closure.WorkflowId.Name, replaceAllInstances)
	response = strings.Replace(response, fmt.Sprintf(substitutionParam, workflowVersion), execution.Closure.WorkflowId.Version, replaceAllInstances)
	response = strings.Replace(response, fmt.Sprintf(substitutionParam, launchPlanProject), execution.Spec.LaunchPlan.Project, replaceAllInstances)
	response = strings.Replace(response, fmt.Sprintf(substitutionParam, launchPlanDomain), execution.Spec.LaunchPlan.Domain, replaceAllInstances)
	response = strings.Replace(response, fmt.Sprintf(substitutionParam, launchPlanName), execution.Spec.LaunchPlan.Name, replaceAllInstances)
	response = strings.Replace(response, fmt.Sprintf(substitutionParam, launchPlanVersion), execution.Spec.LaunchPlan.Version, replaceAllInstances)
	if request.Event.GetError() != nil {
		response = strings.Replace(response, fmt.Sprintf(substitutionParam, errorPlaceholder),
			fmt.Sprintf(executionError, request.Event.GetError().Message), replaceAllInstances)
	} else {
		// Replace the optional error placeholder with an empty string.
		response = strings.Replace(response, fmt.Sprintf(substitutionParam, errorPlaceholder), "", replaceAllInstances)
	}

	return response
}

// Converts a terminal execution event and existing execution model to an admin.EmailMessage proto, substituting parameters
// in customizable email fields set in the flyteadmin application notifications config.
func ToEmailMessageFromWorkflowExecutionEvent(
	config runtimeInterfaces.NotificationsConfig,
	emailNotification admin.EmailNotification,
	request admin.WorkflowExecutionEventRequest,
	execution *admin.Execution) *admin.EmailMessage {

	return &admin.EmailMessage{
		SubjectLine:     substituteEmailParameters(config.NotificationsEmailerConfig.Subject, request, execution),
		SenderEmail:     config.NotificationsEmailerConfig.Sender,
		RecipientsEmail: emailNotification.GetRecipientsEmail(),
		Body:            substituteEmailParameters(config.NotificationsEmailerConfig.Body, request, execution),
	}
}
