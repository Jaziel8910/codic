package codang

// AST nodes for the Codang language.

type Node interface {
	nodeType() string
}

// --- Top level ---

type Program struct {
	Metadata   map[string]string // parsed from @key value lines
	Statements []Node
}

func (*Program) nodeType() string { return "Program" }

// --- Statements ---

type AssignStmt struct {
	Name  string
	Value Node
}

func (*AssignStmt) nodeType() string { return "Assign" }

type ExprStmt struct {
	Expr Node
}

func (*ExprStmt) nodeType() string { return "ExprStmt" }

type FuncDef struct {
	Name   string
	Params []string
	Body   []Node
}

func (*FuncDef) nodeType() string { return "FuncDef" }

type ReturnStmt struct {
	Value Node
}

func (*ReturnStmt) nodeType() string { return "Return" }

type IfStmt struct {
	Cond     Node
	Then     []Node
	ElseBody []Node
}

func (*IfStmt) nodeType() string { return "If" }

// --- Expressions ---

type Ident struct {
	Name string
}

func (*Ident) nodeType() string { return "Ident" }

type NumberLit struct {
	Value float64
}

func (*NumberLit) nodeType() string { return "Number" }

type StringLit struct {
	Value string // mini-notation string
}

func (*StringLit) nodeType() string { return "String" }

type BoolLit struct {
	Value bool
}

func (*BoolLit) nodeType() string { return "Bool" }

type NilLit struct{}

func (*NilLit) nodeType() string { return "Nil" }

type ArrayLit struct {
	Elements []Node
}

func (*ArrayLit) nodeType() string { return "Array" }

// CallExpr is a function call: fn(args...)
type CallExpr struct {
	Callee Node   // can be Ident or MethodCall target
	Name   string // function name (empty if callee is an expression)
	Args   []Node
}

func (*CallExpr) nodeType() string { return "Call" }

// MethodCall is a method call on a value: value.method(args...)
type MethodCall struct {
	Target Node
	Method string
	Args   []Node
}

func (*MethodCall) nodeType() string { return "MethodCall" }

// BinaryOp: left op right
type BinaryOp struct {
	Op    string
	Left  Node
	Right Node
}

func (*BinaryOp) nodeType() string { return "BinOp" }

// UnaryOp: op operand
type UnaryOp struct {
	Op      string
	Operand Node
}

func (*UnaryOp) nodeType() string { return "UnaryOp" }

// Index: value[index]
type Index struct {
	Target Node
	Index  Node
}

func (*Index) nodeType() string { return "Index" }
