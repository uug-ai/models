package properties

// StageCondition remains an alias of WorkflowCondition; preserve its public
// property constants for existing consumers.
const (
	StageConditionPath  = WorkflowConditionPath
	StageConditionOp    = WorkflowConditionOp
	StageConditionValue = WorkflowConditionValue
)
