grammar Workload ;

options { language=Go; }

workloadSteps : workloadStep+ ;
workloadStep : singleStep ;
singleStep: ( repeatStep | dmlStep ) ';' ; 
repeatStep : 'REPEAT' Int '(' singleStep+ ')' ;

dmlStep : insertOp | updateOp | deleteOp | randomDmlOp ;

insertOp : simpleInsertOp | assignmentInsertOp ;
assignmentInsertOp : RowID '=' simpleInsertOp ;
simpleInsertOp : 'INSERT' TableName ;    // insert op can only apply on table

updateOp : simpleUpdateOp | assignmentUpdateOp ;
assignmentUpdateOp : RowID '=' simpleUpdateOp ;
simpleUpdateOp : 'UPDATE' tableOrRowRef ;

deleteOp : simpleDeleteOp ;    // delete op is not assignable
simpleDeleteOp : 'DELETE' tableOrRowRef ;

randomDmlOp : 'RANDOM-DML' TableName ;    //random DML can only apply on table, not assignable

tableOrRowRef : TableName | RowID ;
TableName: UID ( '.' UID ) ? ;
RowID : '@' ID ;
UID : ReverseQuoteID | ID ;
ReverseQuoteID : '`' ~'`'+ '`' ;
Int : [0-9]+ ;
ID : [a-zA-Z_] [a-zA-Z_0-9]* ;
WS : [ \t\r\n]+ -> skip ;
