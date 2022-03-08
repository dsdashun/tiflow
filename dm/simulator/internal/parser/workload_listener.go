// Code generated from Workload.g4 by ANTLR 4.9.3. DO NOT EDIT.

package parser // Workload

import "github.com/antlr/antlr4/runtime/Go/antlr"

// WorkloadListener is a complete listener for a parse tree produced by WorkloadParser.
type WorkloadListener interface {
	antlr.ParseTreeListener

	// EnterWorkloadSteps is called when entering the workloadSteps production.
	EnterWorkloadSteps(c *WorkloadStepsContext)

	// EnterWorkloadStep is called when entering the workloadStep production.
	EnterWorkloadStep(c *WorkloadStepContext)

	// EnterSingleStep is called when entering the singleStep production.
	EnterSingleStep(c *SingleStepContext)

	// EnterRepeatStep is called when entering the repeatStep production.
	EnterRepeatStep(c *RepeatStepContext)

	// EnterDmlStep is called when entering the dmlStep production.
	EnterDmlStep(c *DmlStepContext)

	// EnterInsertOp is called when entering the insertOp production.
	EnterInsertOp(c *InsertOpContext)

	// EnterAssignmentInsertOp is called when entering the assignmentInsertOp production.
	EnterAssignmentInsertOp(c *AssignmentInsertOpContext)

	// EnterSimpleInsertOp is called when entering the simpleInsertOp production.
	EnterSimpleInsertOp(c *SimpleInsertOpContext)

	// EnterUpdateOp is called when entering the updateOp production.
	EnterUpdateOp(c *UpdateOpContext)

	// EnterAssignmentUpdateOp is called when entering the assignmentUpdateOp production.
	EnterAssignmentUpdateOp(c *AssignmentUpdateOpContext)

	// EnterSimpleUpdateOp is called when entering the simpleUpdateOp production.
	EnterSimpleUpdateOp(c *SimpleUpdateOpContext)

	// EnterDeleteOp is called when entering the deleteOp production.
	EnterDeleteOp(c *DeleteOpContext)

	// EnterSimpleDeleteOp is called when entering the simpleDeleteOp production.
	EnterSimpleDeleteOp(c *SimpleDeleteOpContext)

	// EnterRandomDmlOp is called when entering the randomDmlOp production.
	EnterRandomDmlOp(c *RandomDmlOpContext)

	// EnterTableOrRowRef is called when entering the tableOrRowRef production.
	EnterTableOrRowRef(c *TableOrRowRefContext)

	// ExitWorkloadSteps is called when exiting the workloadSteps production.
	ExitWorkloadSteps(c *WorkloadStepsContext)

	// ExitWorkloadStep is called when exiting the workloadStep production.
	ExitWorkloadStep(c *WorkloadStepContext)

	// ExitSingleStep is called when exiting the singleStep production.
	ExitSingleStep(c *SingleStepContext)

	// ExitRepeatStep is called when exiting the repeatStep production.
	ExitRepeatStep(c *RepeatStepContext)

	// ExitDmlStep is called when exiting the dmlStep production.
	ExitDmlStep(c *DmlStepContext)

	// ExitInsertOp is called when exiting the insertOp production.
	ExitInsertOp(c *InsertOpContext)

	// ExitAssignmentInsertOp is called when exiting the assignmentInsertOp production.
	ExitAssignmentInsertOp(c *AssignmentInsertOpContext)

	// ExitSimpleInsertOp is called when exiting the simpleInsertOp production.
	ExitSimpleInsertOp(c *SimpleInsertOpContext)

	// ExitUpdateOp is called when exiting the updateOp production.
	ExitUpdateOp(c *UpdateOpContext)

	// ExitAssignmentUpdateOp is called when exiting the assignmentUpdateOp production.
	ExitAssignmentUpdateOp(c *AssignmentUpdateOpContext)

	// ExitSimpleUpdateOp is called when exiting the simpleUpdateOp production.
	ExitSimpleUpdateOp(c *SimpleUpdateOpContext)

	// ExitDeleteOp is called when exiting the deleteOp production.
	ExitDeleteOp(c *DeleteOpContext)

	// ExitSimpleDeleteOp is called when exiting the simpleDeleteOp production.
	ExitSimpleDeleteOp(c *SimpleDeleteOpContext)

	// ExitRandomDmlOp is called when exiting the randomDmlOp production.
	ExitRandomDmlOp(c *RandomDmlOpContext)

	// ExitTableOrRowRef is called when exiting the tableOrRowRef production.
	ExitTableOrRowRef(c *TableOrRowRefContext)
}
