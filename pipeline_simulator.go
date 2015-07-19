package main

import (
	"fmt"
	// "log"
)

const (
	NOP              int = 0
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
	REG_20_16_MASK   int = 0x001F0000 // >> 16
	REG_15_11_MASK   int = 0x0000F800 // >> 11
	MAX_CLOCK_CYCLES int = 6
)

var Main_Mem = make([]byte, MAX_MEMORY)
var Regs = make([]int, MAX_REGS)
var clock_cyle = 0
var pc = -1

///////////////////
// Instructions //
/////////////////
var instructions = []int{
	0x00a63820,
	0x8d0f0004}

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

////////////////////////
// Pipeline Registers //
///////////////////////
var ifid_w = new(IF_ID_Write)
var ifid_r = new(IF_ID_Read)
var id_ex_w = new(ID_EX_Write)
var id_ex_r = new(ID_EX_Read)
var ex_mem_w = new(EX_MEM_Write)
var ex_mem_r = new(EX_MEM_Read)
var mem_wb_w = new(MEM_WB_Write)
var mem_wb_r = new(MEM_WB_Read)

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

// TODO: do i need this ???
var alu_control_lines = map[string]int{
	"lw":  2,
	"sw":  2,
	"lb":  2,
	"sb":  2,
	"and": 0,
	"or":  1,
	"add": 2,
	"sub": 6,
	"slt": 7,
	"nor": 12}

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
	fmt.Printf("\n")
	fmt.Printf("IF / ID %s \n", r.Reg_Type)
	fmt.Println("------------")
	fmt.Printf("Inst = %.8X     IncrPC = %d \n", r.Instruction, r.Incr_PC)
}

type ID_EX_Base struct {
	RegDst         int
	ALUSrc         int
	ALUOp          int
	MemRead        int
	MemWrite       int
	Branch         int
	MemToReg       int
	RegWrite       int
	Incr_PC        int
	ReadReg1Value  int
	ReadReg2Value  int
	SEOffset       int
	WriteReg_20_16 int
	WriteReg_15_11 int
	Function       int
	Reg_Type       string
	Instr_String   string
	Decoded_Inst   string
}

type ID_EX_Write struct {
	ID_EX_Base // anonymous base type
}

type ID_EX_Read struct {
	ID_EX_Base // anonymous base type
}

func (r *ID_EX_Base) dump_ID_EX() {
	fmt.Printf("\n")
	fmt.Printf("ID / EX %s \n", r.Reg_Type)
	fmt.Println("------------")
	fmt.Printf("Decoded Instruction: %s\n", r.Decoded_Inst)
	fmt.Printf("Control: \n")
	fmt.Printf("RegDst=%d, ALUSrc=%d, ALUOp=%b \n", r.RegDst, r.ALUSrc, r.ALUOp)
	fmt.Printf("MemRead=%d, MemWrite=%d, Branch=%d \n", r.MemRead, r.MemWrite, r.Branch)
	fmt.Printf("MemToReg=%d, RegWrite=%d, [%s] \n", r.MemToReg, r.RegWrite, r.Instr_String)

	fmt.Printf("IncrPC= %d  ReadReg1Value=%X  ReadReg2Value=%X \n", r.Incr_PC, r.ReadReg1Value, r.ReadReg2Value)
	fmt.Printf("SEOffset=%X  WriteReg_20_16=%X  WriteReg_15_11=%X  Function=%X \n", r.SEOffset, r.WriteReg_20_16, r.WriteReg_15_11, r.Function)
}

type EX_MEM_Base struct {
	MemRead      int
	MemWrite     int
	Branch       int
	MemToReg     int
	RegWrite     int
	CalcBTA      int
	Zero         int
	ALUResult    int
	SWValue      int
	WriteRegNum  int
	Reg_Type     string
	Instr_String string
}

type EX_MEM_Write struct {
	EX_MEM_Base // anonymous base type
}

type EX_MEM_Read struct {
	EX_MEM_Base // anonymous base type
}

func (r *EX_MEM_Base) dump_EX_MEM() {
	fmt.Printf("\n")
	fmt.Printf("EX / MEM %s \n", r.Reg_Type)
	fmt.Println("------------")

	fmt.Printf("Control: \n")
	fmt.Printf("MemRead=%d, MemWrite=%d, Branch=%d \n", r.MemRead, r.MemWrite, r.Branch)
	fmt.Printf("MemToReg=%d, RegWrite=%d, [%s] \n", r.MemToReg, r.RegWrite, r.Instr_String)

	fmt.Printf("CalcBTA=%X  Zero=%b  ALUResult=%X \n", r.CalcBTA, r.Zero, r.ALUResult)
	fmt.Printf("SWValue=%X  WriteRegNum=%d \n", r.SWValue, r.WriteRegNum)
}

type MEM_WB_Base struct {
	MemToReg     int
	RegWrite     int
	LWDataValue  int
	ALUResult    int
	WriteRegNum  int
	Reg_Type     string
	Instr_String string
}

type MEM_WB_Write struct {
	MEM_WB_Base // anonymous base type
}

type MEM_WB_Read struct {
	MEM_WB_Base // anonymous base type
}

func (r *MEM_WB_Base) dump_MEM_WB() {
	fmt.Printf("\n")
	fmt.Printf("MEM / WB %s \n", r.Reg_Type)
	fmt.Println("------------")

	fmt.Printf("Control: \n")
	fmt.Printf("MemToReg=%d, RegWrite=%d, [%s] \n", r.MemToReg, r.RegWrite, r.Instr_String)

	fmt.Printf("LWDataValue=%X  ALUResult=%X  WriteRegNum=%d \n", r.LWDataValue, r.ALUResult, r.WriteRegNum)
}

func main() {

	Initialize_Memory()
	Initialize_Registers()
	Initialize_Pipeline()

	Dump_Memory()
	Dump_Registers()

	// dump at Clock-Cycle 0
	Print_out_everything(false)

	for clock_cyle < MAX_CLOCK_CYCLES {
		clock_cyle++

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

	Dump_Memory()
	Dump_Registers()
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

	id_ex_w.Reg_Type = "Write"
	id_ex_r.Reg_Type = "Read"

	ex_mem_w.Reg_Type = "Write"
	ex_mem_r.Reg_Type = "Read"

	mem_wb_w.Reg_Type = "Write"
	mem_wb_r.Reg_Type = "Read"
}

func IF_stage() {
	fetched_inst := NOP
	// carry a tmp pointer in case we arent fetching instructions downstream
	tmp_pc := pc

	// increment the pc
	pc++
	// Fetch Instruction
	if pc < len(instructions) {
		fetched_inst = instructions[pc]
	}

	// reset the pc if the inst is a NOP
	if fetched_inst == NOP {
		pc = tmp_pc
	}

	// Copy to IF/ID Write pipeline register
	ifid_w.Instruction = fetched_inst
	ifid_w.Incr_PC = pc
}

func Get_ALU_Control_Input(inst string) int {
	return alu_control_lines[inst]
}

func ID_stage() {
	showVerbose := false

	// Read instruction from IF/ID Read pipeline
	instruction := ifid_r.Instruction

	// handle opcode format
	if ((instruction & OPCODE_MASK) >> 26) == RFORMAT {
		decoded_inst := Do_RFormat(instruction, showVerbose)

		// Set these for R_Instructions
		// see page 266 for control line settings
		id_ex_w.Incr_PC = ifid_r.Incr_PC
		id_ex_w.ALUOp = 2
		id_ex_w.Function = decoded_inst.funct
		id_ex_w.RegDst = 1
		id_ex_w.RegWrite = 1
		id_ex_w.Instr_String = func_codes[decoded_inst.funct]
		id_ex_w.Decoded_Inst = decoded_inst.inst_string
		id_ex_w.WriteReg_20_16 = (instruction & REG_20_16_MASK) >> 16
		id_ex_w.WriteReg_15_11 = (instruction & REG_15_11_MASK) >> 11

		reg1Value := 0
		reg2Value := 0
		if instruction != NOP {
			reg1Value = Regs[decoded_inst.rs]
			reg2Value = Regs[decoded_inst.rt]
		}

		id_ex_w.ReadReg1Value = reg1Value
		id_ex_w.ReadReg2Value = reg2Value

		// These controls are not set
		id_ex_w.MemRead = 0
		id_ex_w.MemWrite = 0
		id_ex_w.Branch = 0
		id_ex_w.MemToReg = 0
		id_ex_w.SEOffset = 0
		id_ex_w.ALUSrc = 0

	} else {
		decoded_inst := Do_IFormat(instruction, showVerbose)
		// NOTE: see page 260 & 266 & 269 in COD for settings for lw/sw
		id_ex_w.Incr_PC = ifid_r.Incr_PC
		id_ex_w.ALUOp = 0 // IFormat instr are 0 for ALUOp
		id_ex_w.Instr_String = op_codes[decoded_inst.op]
		id_ex_w.Decoded_Inst = decoded_inst.inst_string
		id_ex_w.WriteReg_20_16 = (instruction & REG_20_16_MASK) >> 16
		id_ex_w.WriteReg_15_11 = (instruction & REG_15_11_MASK) >> 11
		id_ex_w.RegDst = 0
		id_ex_w.ALUSrc = 1
		id_ex_w.MemToReg = 1
		id_ex_w.RegWrite = 1
		id_ex_w.MemRead = 1
		id_ex_w.MemWrite = 0
		id_ex_w.Branch = 0

		reg1Value := 0
		reg2Value := 0
		if instruction != NOP {
			reg1Value = Regs[decoded_inst.rs]
			reg2Value = Regs[decoded_inst.rt]
		}

		id_ex_w.ReadReg1Value = reg1Value
		id_ex_w.ReadReg2Value = reg2Value
		id_ex_w.SEOffset = int(decoded_inst.offset)
		id_ex_w.Function = 0 // TODO: not read; could probably be commented out to more accurately simulate register

	}
}

func EX_stage() {
	ex_mem_w.MemRead = id_ex_r.MemRead
	ex_mem_w.MemWrite = id_ex_r.MemWrite
	ex_mem_w.Branch = id_ex_r.Branch
	ex_mem_w.MemToReg = id_ex_r.MemToReg
	ex_mem_w.RegWrite = id_ex_r.RegWrite
	ex_mem_w.Instr_String = id_ex_r.Instr_String

	// TODO: this needs to be paramaterized
	// alu_operation_to_perform := alu_control_lines[id_ex_r.Instr_String]
	result := 0
	// TODO: add support for lw / sw / sub
	if id_ex_r.Instr_String == "add" {
		result = id_ex_r.ReadReg1Value + id_ex_r.ReadReg2Value
	} else if id_ex_r.Instr_String == "sub" {
		result = id_ex_r.ReadReg1Value - id_ex_r.ReadReg2Value
	} else if id_ex_r.Instr_String == "nop" {
		result = 0
	} else if id_ex_r.Instr_String == "lw" {
		// calculate pointer index via taking reg1value + SEOffset
		result = id_ex_r.ReadReg1Value + id_ex_r.SEOffset
	}

	ex_mem_w.ALUResult = result

	if id_ex_r.ALUOp == 2 {
		ex_mem_w.WriteRegNum = id_ex_r.WriteReg_15_11
	} else if id_ex_r.ALUOp == 0 {
		// TODO: double-check this is correct usage of ALUOp
		ex_mem_w.WriteRegNum = id_ex_r.WriteReg_20_16
	}

	ex_mem_w.CalcBTA = 0
	ex_mem_w.Zero = 0 // TODO: when does this get set?
	ex_mem_w.SWValue = id_ex_r.ReadReg2Value
}

func MEM_stage() {
	// TODO: handle lb | lw here
	mem_wb_w.MemToReg = ex_mem_r.MemToReg
	mem_wb_w.RegWrite = ex_mem_r.RegWrite
	mem_wb_w.ALUResult = ex_mem_r.ALUResult
	mem_wb_w.WriteRegNum = ex_mem_r.WriteRegNum
	mem_wb_w.Instr_String = ex_mem_r.Instr_String

	if ex_mem_r.Instr_String == "lw" {
		mem_wb_w.LWDataValue = int(Main_Mem[ex_mem_r.ALUResult])
	} else {
		mem_wb_w.LWDataValue = 0
	}

}

func WB_stage() {
	reg_num := mem_wb_r.WriteRegNum
	reg_val := 0

	// TODO: how is this calculated in a real pipeline?
	if mem_wb_r.Instr_String == "add" || mem_wb_r.Instr_String == "sub" {
		reg_val = mem_wb_r.ALUResult
	} else if mem_wb_r.Instr_String == "lw" {
		reg_val = mem_wb_r.LWDataValue
	}

	Regs[reg_num] = reg_val
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

	id_ex_w.dump_ID_EX()
	id_ex_r.dump_ID_EX()

	ex_mem_w.dump_EX_MEM()
	ex_mem_r.dump_EX_MEM()

	mem_wb_w.dump_MEM_WB()
	mem_wb_r.dump_MEM_WB()

	fmt.Println("\n         --- END ---          \n")
}

func Copy_write_to_read() {
	CopyIFID()
	CopyIDEX()
	CopyEXMEM()
	CopyMEMWB()
}

func CopyIFID() {
	ifid_r.Instruction = ifid_w.Instruction
	ifid_r.Incr_PC = ifid_w.Incr_PC
}

func CopyIDEX() {
	id_ex_r.RegDst = id_ex_w.RegDst
	id_ex_r.ALUSrc = id_ex_w.ALUSrc
	id_ex_r.ALUOp = id_ex_w.ALUOp
	id_ex_r.MemRead = id_ex_w.MemRead
	id_ex_r.MemWrite = id_ex_w.MemWrite
	id_ex_r.Branch = id_ex_w.Branch
	id_ex_r.MemToReg = id_ex_w.MemToReg
	id_ex_r.RegWrite = id_ex_w.RegWrite
	id_ex_r.Incr_PC = id_ex_w.Incr_PC
	id_ex_r.ReadReg1Value = id_ex_w.ReadReg1Value
	id_ex_r.ReadReg2Value = id_ex_w.ReadReg2Value
	id_ex_r.SEOffset = id_ex_w.SEOffset
	id_ex_r.WriteReg_20_16 = id_ex_w.WriteReg_20_16
	id_ex_r.WriteReg_15_11 = id_ex_w.WriteReg_15_11
	id_ex_r.Function = id_ex_w.Function
	id_ex_r.Instr_String = id_ex_w.Instr_String
}

func CopyEXMEM() {
	ex_mem_r.MemRead = ex_mem_w.MemRead
	ex_mem_r.MemWrite = ex_mem_w.MemWrite
	ex_mem_r.Branch = ex_mem_w.Branch
	ex_mem_r.MemToReg = ex_mem_w.MemToReg
	ex_mem_r.RegWrite = ex_mem_w.RegWrite
	ex_mem_r.CalcBTA = ex_mem_w.CalcBTA
	ex_mem_r.Zero = ex_mem_w.Zero
	ex_mem_r.ALUResult = ex_mem_w.ALUResult
	ex_mem_r.SWValue = ex_mem_w.SWValue
	ex_mem_r.WriteRegNum = ex_mem_w.WriteRegNum
	ex_mem_r.Instr_String = ex_mem_w.Instr_String
}

func CopyMEMWB() {
	mem_wb_r.MemToReg = mem_wb_w.MemToReg
	mem_wb_r.RegWrite = mem_wb_w.RegWrite
	mem_wb_r.LWDataValue = mem_wb_w.LWDataValue
	mem_wb_r.ALUResult = mem_wb_w.ALUResult
	mem_wb_r.WriteRegNum = mem_wb_w.WriteRegNum
	mem_wb_r.Instr_String = mem_wb_w.Instr_String
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

func Dump_Memory() {
	// TODO: dump all memory ?
	fmt.Printf("Main_Mem[0x10C]=[%X]\n", Main_Mem[0x10C])
}

func Dump_Registers() {
	fmt.Printf("Registers: %X\n", Regs)
}
