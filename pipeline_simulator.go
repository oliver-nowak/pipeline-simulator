package main

import (
	"fmt"
)

const (
	MAX_MEMORY       int = 1024
	MAX_REGS         int = 32
	INSTRUCTION_SIZE int = 0x00000004 // in bytes
	RFORMAT          int = 0x0
	OPCODE_MASK      int = 0xFC000000 // >> 26
	RS_MASK          int = 0x03E00000 // >> 21
	RT_MASK          int = 0x001F0000 // >> 16
	RD_MASK          int = 0x0000F800 // >> 11
	FUNC_MASK        int = 0x0000003F // >> 00
	OFFSET_MASK      int = 0x0000FFFF // >> 00
)

var Main_Mem = make([]byte, MAX_MEMORY)
var Regs = make([]int, MAX_REGS)
var clock_cyle = 0
var pc = 0
var exit_requested = false

///////////////////
// Instructions //
/////////////////
var instructions = []int{
	0x00a63820}

// 0x8d0f0004,
// 0xad09fffc,
// 0x00625022}

// TODO: swap for assignment instructions
// var instructions = []int{
// 	0xa1020000,
// 	0x810AFFFC,
// 	0x00831820,
// 	0x01263820,
// 	0x01224820,
// 	0x81180000,
// 	0x81510010,
// 	0x00624022,
// 	0x00000000,
// 	0x00000000,
// 	0x00000000,
// 	0x00000000}

var ifid_w = new(IF_ID_Write)
var ifid_r = new(IF_ID_Read)

// TODO: remove inst codes not requested by assignment parameters
var func_codes = map[int]string{
	0x0:  "nop",
	0x20: "add",
	0x22: "sub"}

var op_codes = map[int]string{
	0x20: "lb",
	0x23: "lw",
	0x28: "sb",
	0x2b: "sw"}

type Base_Inst struct {
	instruction int
	inst_string string
	code        string
}

type R_Inst struct {
	Base_Inst         // anonymous base type
	funct, rd, rs, rt int
}

type I_Inst struct {
	Base_Inst  // anonymous base type
	op, rt, rs int
	offset     int16
}

type IF_ID_Base struct {
	Instruction        int
	Code               string
	Incr_PC            int
	Reg_Type           string
	Instruction_String string
}

type IF_ID_Write struct {
	IF_ID_Base // anonymous base type
}

type IF_ID_Read struct {
	IF_ID_Base // anonymous base type
}

func (r *IF_ID_Base) dump_IF_ID() {
	fmt.Printf("\n")
	fmt.Printf("IF / ID %s \n", r.Reg_Type)
	fmt.Println("------------")
	fmt.Printf("Inst = %.8X     [%s]     IncrPC = %d \n", r.Instruction, r.Instruction_String, r.Incr_PC)
}

type ID_EX_Base struct {
	RegDst         bool
	ALUSrc         bool
	ALUOp          bool
	MemRead        bool
	MemWrite       bool
	Branch         bool
	MemToReg       bool
	RegWrite       bool
	Incr_PC        int
	ReadReg1Value  int
	ReadReg2Value  int
	SEOffset       int
	WriteReg_20_16 int
	WriteReg_15_11 int
	Function       int
	Reg_Type       string
	Instr_String   string
}

type ID_EX_Write struct {
	ID_EX_Base
}

type ID_EX_Read struct {
	ID_EX_Base
}

func (r *ID_EX_Base) dump_ID_EX() {
	fmt.Printf("\n")
	fmt.Printf("ID / EX %s \n", r.Reg_Type)
	fmt.Println("------------")

	fmt.Printf("Control: \n")
	fmt.Printf("RegDst=%d, ALUSrc=%d, ALUOp=%d", r.RegDst, r.ALUSrc, r.ALUOp)
	fmt.Printf("MemRead=%d, MemWrite=%d, Branch=%d", r.MemRead, r.MemWrite, r.Branch)
	fmt.Printf("MemToReg=%d, RegWrite=%d, [%s] \n", r.MemToReg, r.RegWrite, r.Instr_String)

	fmt.Printf("IncrPC= %d  ReadReg1Value=%X  ReadReg2Value=%X", r.Incr_PC, r.ReadReg1Value, r.ReadReg2Value)
	fmt.Printf("SEOffset=%X  WriteReg_20_16=%X  WriteReg_15_11=%X  Function=%d", r.SEOffset, r.WriteReg_20_16, r.WriteReg_15_11, r.Function)
}

type EX_MEM_Base struct {
	MemRead      bool
	MemWrite     bool
	Branch       bool
	MemToReg     bool
	RegWrite     bool
	CalcBTA      int
	Zero         bool
	ALUResult    int
	SWValue      int
	WriteRegNum  int
	Reg_Type     string
	Instr_String string
}

type EX_MEM_Write struct {
	EX_MEM_Base
}

type EX_MEM_Read struct {
	EX_MEM_Base
}

func (r *EX_MEM_Base) dump_EX_MEM() {
	fmt.Printf("\n")
	fmt.Printf("EX / MEM %s \n", r.Reg_Type)
	fmt.Println("------------")

	fmt.Printf("Control: \n")
	fmt.Printf("MemRead=%d, MemWrite=%d, Branch=%d", r.MemRead, r.MemWrite, r.Branch)
	fmt.Printf("MemToReg=%d, RegWrite=%d, [%s] \n", r.MemToReg, r.RegWrite, r.Instr_String)

	fmt.Printf("CalcBTA=%X  Zero=%b  ALUResult=%X", r.CalcBTA, r.Zero, r.ALUResult)
	fmt.Printf("SWValue=%X  WriteRegNum=%d", r.SWValue, r.WriteRegNum)
}

type MEM_WB_Base struct {
	MemToReg     bool
	RegWrite     bool
	LWDataValue  int
	ALUResult    int
	WriteRegNum  int
	Reg_Type     string
	Instr_String string
}

type MEM_WB_Write struct {
	MEM_WB_Base
}

type MEM_WB_Read struct {
	MEM_WB_Base
}

func (r *MEM_WB_Base) dump_MEM_WB() {
	fmt.Printf("\n")
	fmt.Printf("MEM / WB %s \n", r.Reg_Type)
	fmt.Println("------------")

	fmt.Printf("Control: \n")
	fmt.Printf("MemToReg=%d, RegWrite=%d, [%s] \n", r.MemToReg, r.RegWrite, r.Instr_String)

	fmt.Printf("LWDataValue=%X  ALUResult=%X  WriteRegNum=%d", r.LWDataValue, r.ALUResult, r.WriteRegNum)
}

func main() {

	Initialize_Memory()
	Initialize_Registers()
	Initialize_Pipeline()

	// fmt.Printf("Main_Mem[0x101]=[%X]\n", Main_Mem[0x101])
	// fmt.Printf("Registers: [%X]\n", Regs)

	// dump at Clock-Cycle 0
	Print_out_everything(false)

	for !exit_requested {
		IF_stage()
		ID_stage()
		EX_stage()
		MEM_stage()
		WB_stage()
		Print_out_everything(false)
		Copy_write_to_read()

		// check write pipeline registers
		// TODO: remove this call, and remove param from signature
		Print_out_everything(true)
	}
}

func Initialize_Memory() {
	// Initialize Main Memory so that Address 0x100 = byte value 00, and so on.
	max := 256
	mem_val := -1
	for i := range Main_Mem {
		mem_val = (mem_val + 1) % max
		Main_Mem[i] = byte(mem_val)
	}
}

func Initialize_Registers() {
	for i := range Regs {
		Regs[i] = 0x100 + i
	}
}

func Initialize_Pipeline() {
	ifid_w.Reg_Type = "Write"
	ifid_r.Reg_Type = "Read"
}

func IF_stage() {
	// Fetch Instruction
	clock_cyle++
	fetched_inst := instructions[pc]
	pc++

	// Copy to IF/ID Write pipeline register
	ifid_w.Instruction = fetched_inst
	ifid_w.Incr_PC = pc
}

func ID_stage() {
	showVerbose := false

	// Read instruction from IF/ID Read pipeline
	instruction := ifid_r.Instruction

	// handle opcode format
	if ((instruction & OPCODE_MASK) >> 26) == RFORMAT {
		decoded_inst := Do_RFormat(instruction, showVerbose)

	} else {
		decoded_inst := Do_IFormat(instruction, showVerbose)
	}
}

func EX_stage() {

}

func MEM_stage() {

}

func WB_stage() {

}

func Print_out_everything(isAfterCopy bool) {
	copyFlag := "Before"
	if isAfterCopy {
		copyFlag = "After"
	}

	fmt.Println("\n")
	fmt.Printf("Clock Cycle %d [%s]\n", clock_cyle, copyFlag)
	fmt.Println("------------------------------")
	ifid_w.dump_IF_ID()
	ifid_r.dump_IF_ID()

	fmt.Println("         --- END ---          \n")
}

func Copy_write_to_read() {
	CopyIFID()
}

func CopyIFID() {
	ifid_r.Instruction = ifid_w.Instruction
	ifid_r.Incr_PC = ifid_r.Incr_PC
}
func Do_RFormat(instruction int, showVerbose bool) *R_Inst {
	opcode := (instruction & OPCODE_MASK) >> 26
	rs := (instruction & RS_MASK) >> 21
	rt := (instruction & RT_MASK) >> 16
	rd := (instruction & RD_MASK) >> 11
	funct := (instruction & FUNC_MASK) >> 0
	func_code := func_codes[funct]

	if showVerbose {
		fmt.Println("---RFORMAT---")
		fmt.Printf("Instruction : 0x%X \n", instruction)
		fmt.Printf("Opcode      : 0x%X \n", opcode)
		fmt.Printf("RS_MASK          : 0x%X \n", rs)
		fmt.Printf("RT_MASK          : 0x%X \n", rt)
		fmt.Printf("RD_MASK          : 0x%X \n", rd)
		fmt.Printf("Func        : 0x%X \n", funct)
		fmt.Println("---END--")
	}

	inst_string := fmt.Sprintf("%3s  $%d, $%d, $%d", func_code, rd, rs, rt)
	// fmt.Println(inst_string)

	r := new(R_Inst)
	r.instruction = instruction
	r.funct = funct
	r.rd = rd
	r.rs = rs
	r.rt = rt
	r.inst_string = inst_string

	return r
}

func Do_IFormat(instruction int, showVerbose bool) *I_Inst {
	opcode := (instruction & OPCODE_MASK) >> 26
	op := op_codes[opcode]
	rs := (instruction & RS_MASK) >> 21
	rt := (instruction & RT_MASK) >> 16
	// cast to int16 in order to get correct signed number
	offset := int16((instruction & OFFSET_MASK))
	decompressed_offset := offset << 2

	if showVerbose {
		fmt.Println("---IFORMAT---")
		fmt.Printf("Instruction : 0x%X \n", instruction)
		fmt.Printf("Opcode      : [0x%X] %s \n", opcode, op)
		fmt.Printf("RS_MASK          : 0x%X \n", rs)
		fmt.Printf("RT_MASK          : 0x%X \n", rt)
		fmt.Printf("Offset      : 0x%X \n", offset)
		fmt.Printf("D-Offset    : 0x%X \n", decompressed_offset)
		fmt.Println("---END---")
	}

	inst_string := fmt.Sprintf("%3s  $%d, %d($%d)", op, rt, offset, rs)

	// TODO: remove this before handing in assignment
	// if op == "lb" || op == "sb" {
	// fmt.Println(inst_string)
	// } else {
	// fmt.Printf("Unknown I_Inst: %3s  $%d,	$%d,	address %X \n", op, rt, rs, offset)
	// log.Fatal("Leaving.")
	// }

	i := new(I_Inst)
	i.instruction = instruction
	i.op = opcode
	i.rt = rt
	i.offset = offset
	i.rs = rs
	i.inst_string = inst_string

	return i
}
