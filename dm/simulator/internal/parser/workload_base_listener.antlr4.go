// Code generated from ../../grammar/Workload.g4 by ANTLR 4.9.3. DO NOT EDIT.

package parser // Workload
import "github.com/antlr/antlr4/runtime/Go/antlr"

// BaseWorkloadListener is a complete listener for a parse tree produced by WorkloadParser.
type BaseWorkloadListener struct{}

var _ WorkloadListener = &BaseWorkloadListener{}

// VisitTerminal is called when a terminal node is visited.
func (s *BaseWorkloadListener) VisitTerminal(node antlr.TerminalNode) {}

// VisitErrorNode is called when an error node is visited.
func (s *BaseWorkloadListener) VisitErrorNode(node antlr.ErrorNode) {}

// EnterEveryRule is called when any rule is entered.
func (s *BaseWorkloadListener) EnterEveryRule(ctx antlr.ParserRuleContext) {}

// ExitEveryRule is called when any rule is exited.
func (s *BaseWorkloadListener) ExitEveryRule(ctx antlr.ParserRuleContext) {}

// EnterWorkloadSteps is called when production workloadSteps is entered.
func (s *BaseWorkloadListener) EnterWorkloadSteps(ctx *WorkloadStepsContext) {}

// ExitWorkloadSteps is called when production workloadSteps is exited.
func (s *BaseWorkloadListener) ExitWorkloadSteps(ctx *WorkloadStepsContext) {}

// EnterWorkloadStep is called when production workloadStep is entered.
func (s *BaseWorkloadListener) EnterWorkloadStep(ctx *WorkloadStepContext) {}

// ExitWorkloadStep is called when production workloadStep is exited.
func (s *BaseWorkloadListener) ExitWorkloadStep(ctx *WorkloadStepContext) {}

// EnterSingleStep is called when production singleStep is entered.
func (s *BaseWorkloadListener) EnterSingleStep(ctx *SingleStepContext) {}

// ExitSingleStep is called when production singleStep is exited.
func (s *BaseWorkloadListener) ExitSingleStep(ctx *SingleStepContext) {}

// EnterRepeatStep is called when production repeatStep is entered.
func (s *BaseWorkloadListener) EnterRepeatStep(ctx *RepeatStepContext) {}

// ExitRepeatStep is called when production repeatStep is exited.
func (s *BaseWorkloadListener) ExitRepeatStep(ctx *RepeatStepContext) {}

// EnterDmlStep is called when production dmlStep is entered.
func (s *BaseWorkloadListener) EnterDmlStep(ctx *DmlStepContext) {}

// ExitDmlStep is called when production dmlStep is exited.
func (s *BaseWorkloadListener) ExitDmlStep(ctx *DmlStepContext) {}

// EnterInsertOp is called when production insertOp is entered.
func (s *BaseWorkloadListener) EnterInsertOp(ctx *InsertOpContext) {}

// ExitInsertOp is called when production insertOp is exited.
func (s *BaseWorkloadListener) ExitInsertOp(ctx *InsertOpContext) {}

// EnterAssignmentInsertOp is called when production assignmentInsertOp is entered.
func (s *BaseWorkloadListener) EnterAssignmentInsertOp(ctx *AssignmentInsertOpContext) {}

// ExitAssignmentInsertOp is called when production assignmentInsertOp is exited.
func (s *BaseWorkloadListener) ExitAssignmentInsertOp(ctx *AssignmentInsertOpContext) {}

// EnterSimpleInsertOp is called when production simpleInsertOp is entered.
func (s *BaseWorkloadListener) EnterSimpleInsertOp(ctx *SimpleInsertOpContext) {}

// ExitSimpleInsertOp is called when production simpleInsertOp is exited.
func (s *BaseWorkloadListener) ExitSimpleInsertOp(ctx *SimpleInsertOpContext) {}

// EnterUpdateOp is called when production updateOp is entered.
func (s *BaseWorkloadListener) EnterUpdateOp(ctx *UpdateOpContext) {}

// ExitUpdateOp is called when production updateOp is exited.
func (s *BaseWorkloadListener) ExitUpdateOp(ctx *UpdateOpContext) {}

// EnterAssignmentUpdateOp is called when production assignmentUpdateOp is entered.
func (s *BaseWorkloadListener) EnterAssignmentUpdateOp(ctx *AssignmentUpdateOpContext) {}

// ExitAssignmentUpdateOp is called when production assignmentUpdateOp is exited.
func (s *BaseWorkloadListener) ExitAssignmentUpdateOp(ctx *AssignmentUpdateOpContext) {}

// EnterSimpleUpdateOp is called when production simpleUpdateOp is entered.
func (s *BaseWorkloadListener) EnterSimpleUpdateOp(ctx *SimpleUpdateOpContext) {}

// ExitSimpleUpdateOp is called when production simpleUpdateOp is exited.
func (s *BaseWorkloadListener) ExitSimpleUpdateOp(ctx *SimpleUpdateOpContext) {}

// EnterDeleteOp is called when production deleteOp is entered.
func (s *BaseWorkloadListener) EnterDeleteOp(ctx *DeleteOpContext) {}

// ExitDeleteOp is called when production deleteOp is exited.
func (s *BaseWorkloadListener) ExitDeleteOp(ctx *DeleteOpContext) {}

// EnterSimpleDeleteOp is called when production simpleDeleteOp is entered.
func (s *BaseWorkloadListener) EnterSimpleDeleteOp(ctx *SimpleDeleteOpContext) {}

// ExitSimpleDeleteOp is called when production simpleDeleteOp is exited.
func (s *BaseWorkloadListener) ExitSimpleDeleteOp(ctx *SimpleDeleteOpContext) {}

// EnterRandomDmlOp is called when production randomDmlOp is entered.
func (s *BaseWorkloadListener) EnterRandomDmlOp(ctx *RandomDmlOpContext) {}

// ExitRandomDmlOp is called when production randomDmlOp is exited.
func (s *BaseWorkloadListener) ExitRandomDmlOp(ctx *RandomDmlOpContext) {}

// EnterTableOrRowRef is called when production tableOrRowRef is entered.
func (s *BaseWorkloadListener) EnterTableOrRowRef(ctx *TableOrRowRefContext) {}

// ExitTableOrRowRef is called when production tableOrRowRef is exited.
func (s *BaseWorkloadListener) ExitTableOrRowRef(ctx *TableOrRowRefContext) {}
