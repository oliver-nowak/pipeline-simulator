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

///////////////////
// Instructions //
/////////////////
var instructions = []int{
	0x00a63820}

// 0x8d0f0004,
// 0xad09fffc,
// 0x00625022}

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

var func_codes = map[int]string{
	0x0:  "nop",
	0x20: "add",
	0x22: "sub"}

var op_codes = map[int]string{
	0x20: "lb",
	0x23: "lw",
	0x28: "sb",
	0x2b: "sw"}

type R_Inst struct {
	instruction       int
	funct, rd, rs, rt int
	code              string
	inst_string       string
}

type I_Inst struct {
	instruction int
	op, rt, rs  int
	offset      int16
	code        string
	inst_string string
}

type IF_ID_Base struct {
	Instruction int
	Code        string
	Incr_PC     int
	Reg_Type    string
}

type IF_ID_Write struct {
	IF_ID_Base // anonymous base type
}

type IF_ID_Read struct {
	IF_ID_Base // anonymous base type
}

func (r *IF_ID_Base) dump_IF_ID() {
	fmt.Printf("IF / ID %s \n", r.Reg_Type)
	fmt.Println("------------")
	fmt.Printf("Inst = %X     [%s]     IncrPC = %d \n", r.Instruction, r.Code, r.Incr_PC)
}

type ID_EX_Write struct {
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
}

type ID_EX_Read struct {
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
}

type EX_MEM_Write struct {
	MemRead     bool
	MemWrite    bool
	Branch      bool
	MemToReg    bool
	RegWrite    bool
	CalcBTA     int
	Zero        bool
	ALUResult   int
	SWValue     int
	WriteRegNum int
}

type EX_MEM_Read struct {
	MemRead     bool
	MemWrite    bool
	Branch      bool
	MemToReg    bool
	RegWrite    bool
	CalcBTA     int
	Zero        bool
	ALUResult   int
	SWValue     int
	WriteRegNum int
}

type MEM_WB_Write struct {
	MemToReg    bool
	RegWrite    bool
	LWDataValue int
	ALUResult   int
	WriteRegNum int
}

type MEM_WB_Read struct {
	MemToReg    bool
	RegWrite    bool
	LWDataValue int
	ALUResult   int
	WriteRegNum int
}

func main() {

	Initialize_Memory()
	Initialize_Registers()
	Initialize_Pipeline()

	fmt.Printf("Main_Mem[0x101]=[%X]\n", Main_Mem[0x101])
	fmt.Printf("Registers: [%X]\n", Regs)

	Print_out_everything()

	for pc, instruction := range instructions {
		IF_stage(pc, instruction)
		ID_stage()
		EX_stage()
		MEM_stage()
		WB_stage()
		Print_out_everything()
		Copy_write_to_read()
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

func IF_stage(pc int, instruction int) {
	showVerbose := false

	// handle opcode format
	if ((instruction & OPCODE_MASK) >> 26) == RFORMAT {
		Do_RFormat(instruction, showVerbose)
	} else {
		Do_IFormat(instruction, showVerbose)
	}
}

func ID_stage() {

}

func EX_stage() {

}

func MEM_stage() {

}

func WB_stage() {

}

func Print_out_everything() {
	fmt.Printf("Clock Cycle %d \n", clock_cyle)
	fmt.Println("------------------------------")
	ifid_w.dump_IF_ID()

	fmt.Println("         --- END ---          ")
}

func Copy_write_to_read() {

}

func Do_RFormat(instruction int, showVerbose bool) R_Inst {
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

	inst_string := fmt.Sprintf("[%3s  $%d, $%d, $%d]", func_code, rd, rs, rt)
	fmt.Println(inst_string)

	return R_Inst{instruction: instruction, funct: funct, rd: rd, rs: rs, rt: rt, inst_string: inst_string}
}

func Do_IFormat(instruction int, showVerbose bool) I_Inst {
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

	inst_string := fmt.Sprintf("[%3s  $%d, %d($%d)]", op, rt, offset, rs)

	// if op == "lb" || op == "sb" {
	fmt.Println(inst_string)
	// } else {
	// fmt.Printf("Unknown I_Inst: %3s  $%d,	$%d,	address %X \n", op, rt, rs, offset)
	// log.Fatal("Leaving.")
	// }

	return I_Inst{instruction: instruction, op: opcode, rt: rt, offset: offset, rs: rs, inst_string: inst_string}
}
