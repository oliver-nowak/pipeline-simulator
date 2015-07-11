package main

import (
	"fmt"
	"log"
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

var instructions = []int{
	0xa1020000,
	0x810AFFFC,
	0x00831820,
	0x01263820,
	0x01224820,
	0x81180000,
	0x81510010,
	0x00624022,
	0x00000000,
	0x00000000,
	0x00000000,
	0x00000000}

// var instructions = []int{
// 	0x022DA822, // sub
// 	0x8EF30018, // lw
// 	0x12A70004, // beq
// 	0x02689820,
// 	0xAD930018,
// 	0x02697824,
// 	0xAD8FFFF4,
// 	0x018C6020,
// 	0x02A4A825,
// 	0x158FFFF6,
// 	0x8E59FFF0}

// var func_codes = map[int]string{
// 	0x20: "add",
// 	0x22: "sub",
// 	0x24: "and",
// 	0x25: "or",
// 	0x2A: "slt"}

// var op_codes = map[int]string{
// 	0x04: "beq",
// 	0x05: "bne",
// 	0x23: "lw",
// 	0x2B: "sw"}

// var pc int = 0x7A05C

var func_codes = map[int]string{
	0x0:  "nop",
	0x20: "add",
	0x22: "sub"}

var op_codes = map[int]string{
	0x20: "lb",
	0x28: "sb"}

type R_Inst struct {
	instruction       int
	funct, rd, rs, rt int
}

type I_Inst struct {
	instruction int
	op, rt, rs  int
	offset      int16
}

func main() {

	Initialize_Memory()
	Initialize_Registers()

	fmt.Printf("Main_Mem[0x101]=[%X]\n", Main_Mem[0x101])
	fmt.Printf("Registers: [%X]\n", Regs)

	IF_stage()

	// num_instructions := len(instructions)

	// for num_instructions >= 0 {

	// }
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

func IF_stage() {
	showVerbose := false

	for _, instruction := range instructions {

		// increment pc
		// pc = pc + 0x4

		// handle opcode format
		if ((instruction & OPCODE_MASK) >> 26) == RFORMAT {
			Do_RFormat(instruction, showVerbose)
		} else {
			Do_IFormat(instruction, showVerbose)
		}
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

	fmt.Printf("R_Inst: %3s  $%d,	$%d,	$%d \n", func_code, rd, rs, rt)

	return R_Inst{instruction: instruction, funct: funct, rd: rd, rs: rs, rt: rt}
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

	if op == "lb" || op == "sb" {
		fmt.Printf("I_Inst: %3s  $%d,	%d($%d) \n", op, rt, offset, rs)
	} else {
		fmt.Printf("Unknown I_Inst: %3s  $%d,	$%d,	address %X \n", op, rt, rs, offset)
		log.Fatal("Leaving.")
	}

	return I_Inst{instruction: instruction, op: opcode, rt: rt, offset: offset, rs: rs}
}
