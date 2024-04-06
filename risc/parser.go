package risc

import (
	"fmt"
	"strconv"
	"strings"
)

func Parse(s string) (Application, error) {
	var instructions []InstructionRunner
	labels := make(map[string]int32)
	var pc int32

	for _, line := range strings.Split(s, "\n") {
		line = strings.TrimSpace(line)
		if len(line) == 0 || line[0] == '#' {
			continue
		}

		firstWhitespace := strings.Index(line, " ")
		lastCharacters := line[len(line)-1]
		if firstWhitespace == -1 && lastCharacters != ':' {
			return Application{}, fmt.Errorf("invalid line: %s", line)
		} else if firstWhitespace == -1 && lastCharacters == ':' {
			labels[line[:len(line)-1]] = pc
			continue
		}

		remainingLine := line[firstWhitespace+1:]
		comment := strings.Index(remainingLine, "#")
		if comment != -1 {
			remainingLine = strings.TrimSpace(remainingLine[:comment])
		}

		elements := strings.Split(remainingLine, ",")

		switch strings.ToLower(line[:firstWhitespace]) {
		case "add":
			if err := validateArgs(3, elements, remainingLine); err != nil {
				return Application{}, fmt.Errorf("line %s: %v", remainingLine, err)
			}
			rd, err := parseRegister(strings.TrimSpace(elements[0]))
			if err != nil {
				return Application{}, fmt.Errorf("line %s: %v", remainingLine, err)
			}
			rs1, err := parseRegister(strings.TrimSpace(elements[1]))
			if err != nil {
				return Application{}, fmt.Errorf("line %s: %v", remainingLine, err)
			}
			rs2, err := parseRegister(strings.TrimSpace(elements[2]))
			if err != nil {
				return Application{}, fmt.Errorf("line %s: %v", remainingLine, err)
			}
			instructions = append(instructions, &add{
				rd:  rd,
				rs1: rs1,
				rs2: rs2,
			})
		case "and":
			if err := validateArgs(3, elements, remainingLine); err != nil {
				return Application{}, fmt.Errorf("line %s: %v", remainingLine, err)
			}
			rd, err := parseRegister(strings.TrimSpace(elements[0]))
			if err != nil {
				return Application{}, fmt.Errorf("line %s: %v", remainingLine, err)
			}
			rs1, err := parseRegister(strings.TrimSpace(elements[1]))
			if err != nil {
				return Application{}, fmt.Errorf("line %s: %v", remainingLine, err)
			}
			rs2, err := parseRegister(strings.TrimSpace(elements[2]))
			if err != nil {
				return Application{}, fmt.Errorf("line %s: %v", remainingLine, err)
			}
			instructions = append(instructions, &and{
				rd:  rd,
				rs1: rs1,
				rs2: rs2,
			})
		case "addi":
			if err := validateArgs(3, elements, remainingLine); err != nil {
				return Application{}, fmt.Errorf("line %s: %v", remainingLine, err)
			}
			rd, err := parseRegister(strings.TrimSpace(elements[0]))
			if err != nil {
				return Application{}, fmt.Errorf("line %s: %v", remainingLine, err)
			}
			rs, err := parseRegister(strings.TrimSpace(elements[1]))
			if err != nil {
				return Application{}, fmt.Errorf("line %s: %v", remainingLine, err)
			}
			imm, err := strconv.ParseInt(strings.TrimSpace(elements[2]), 10, 32)
			if err != nil {
				return Application{}, fmt.Errorf("line %s: %v", remainingLine, err)
			}
			instructions = append(instructions, &addi{
				imm: int32(imm),
				rd:  rd,
				rs:  rs,
			})
		case "andi":
			if err := validateArgs(3, elements, remainingLine); err != nil {
				return Application{}, fmt.Errorf("line %s: %v", remainingLine, err)
			}
			rd, err := parseRegister(strings.TrimSpace(elements[0]))
			if err != nil {
				return Application{}, fmt.Errorf("line %s: %v", remainingLine, err)
			}
			rs, err := parseRegister(strings.TrimSpace(elements[1]))
			if err != nil {
				return Application{}, fmt.Errorf("line %s: %v", remainingLine, err)
			}
			imm, err := strconv.ParseInt(strings.TrimSpace(elements[2]), 10, 32)
			if err != nil {
				return Application{}, fmt.Errorf("line %s: %v", remainingLine, err)
			}
			instructions = append(instructions, &andi{
				imm: int32(imm),
				rd:  rd,
				rs:  rs,
			})
		case "auipc":
			if err := validateArgs(2, elements, remainingLine); err != nil {
				return Application{}, fmt.Errorf("line %s: %v", remainingLine, err)
			}
			rd, err := parseRegister(strings.TrimSpace(elements[0]))
			if err != nil {
				return Application{}, fmt.Errorf("line %s: %v", remainingLine, err)
			}
			imm, err := strconv.ParseInt(strings.TrimSpace(elements[1]), 10, 32)
			if err != nil {
				return Application{}, fmt.Errorf("line %s: %v", remainingLine, err)
			}
			instructions = append(instructions, &auipc{
				rd:  rd,
				imm: int32(imm),
			})
		case "beq":
			if err := validateArgs(3, elements, remainingLine); err != nil {
				return Application{}, fmt.Errorf("line %s: %v", remainingLine, err)
			}
			rd, err := parseRegister(strings.TrimSpace(elements[0]))
			if err != nil {
				return Application{}, fmt.Errorf("line %s: %v", remainingLine, err)
			}
			rs, err := parseRegister(strings.TrimSpace(elements[1]))
			if err != nil {
				return Application{}, fmt.Errorf("line %s: %v", remainingLine, err)
			}
			label := strings.TrimSpace(elements[2])
			instructions = append(instructions, &beq{
				rs1:   rd,
				rs2:   rs,
				label: label,
			})
		case "beqz":
			if err := validateArgs(2, elements, remainingLine); err != nil {
				return Application{}, fmt.Errorf("line %s: %v", remainingLine, err)
			}
			rs, err := parseRegister(strings.TrimSpace(elements[0]))
			if err != nil {
				return Application{}, fmt.Errorf("line %s: %v", remainingLine, err)
			}
			label := strings.TrimSpace(elements[1])
			instructions = append(instructions, &beqz{
				rs:    rs,
				label: label,
			})
		case "bge":
			if err := validateArgs(3, elements, remainingLine); err != nil {
				return Application{}, fmt.Errorf("line %s: %v", remainingLine, err)
			}
			rs1, err := parseRegister(strings.TrimSpace(elements[0]))
			if err != nil {
				return Application{}, fmt.Errorf("line %s: %v", remainingLine, err)
			}
			rs2, err := parseRegister(strings.TrimSpace(elements[1]))
			if err != nil {
				return Application{}, fmt.Errorf("line %s: %v", remainingLine, err)
			}
			label := strings.TrimSpace(elements[2])
			instructions = append(instructions, &bge{
				rs1:   rs1,
				rs2:   rs2,
				label: label,
			})
		case "bgeu":
			if err := validateArgs(3, elements, remainingLine); err != nil {
				return Application{}, fmt.Errorf("line %s: %v", remainingLine, err)
			}
			rs1, err := parseRegister(strings.TrimSpace(elements[0]))
			if err != nil {
				return Application{}, fmt.Errorf("line %s: %v", remainingLine, err)
			}
			rs2, err := parseRegister(strings.TrimSpace(elements[1]))
			if err != nil {
				return Application{}, fmt.Errorf("line %s: %v", remainingLine, err)
			}
			label := strings.TrimSpace(elements[2])
			instructions = append(instructions, &bgeu{
				rs1:   rs1,
				rs2:   rs2,
				label: label,
			})
		case "blt":
			if err := validateArgs(3, elements, remainingLine); err != nil {
				return Application{}, fmt.Errorf("line %s: %v", remainingLine, err)
			}
			rs1, err := parseRegister(strings.TrimSpace(elements[0]))
			if err != nil {
				return Application{}, fmt.Errorf("line %s: %v", remainingLine, err)
			}
			rs2, err := parseRegister(strings.TrimSpace(elements[1]))
			if err != nil {
				return Application{}, fmt.Errorf("line %s: %v", remainingLine, err)
			}
			label := strings.TrimSpace(elements[2])
			instructions = append(instructions, &blt{
				rs1:   rs1,
				rs2:   rs2,
				label: label,
			})
		case "bltu":
			if err := validateArgs(3, elements, remainingLine); err != nil {
				return Application{}, fmt.Errorf("line %s: %v", remainingLine, err)
			}
			rs1, err := parseRegister(strings.TrimSpace(elements[0]))
			if err != nil {
				return Application{}, fmt.Errorf("line %s: %v", remainingLine, err)
			}
			rs2, err := parseRegister(strings.TrimSpace(elements[1]))
			if err != nil {
				return Application{}, fmt.Errorf("line %s: %v", remainingLine, err)
			}
			label := strings.TrimSpace(elements[2])
			instructions = append(instructions, &bltu{
				rs1:   rs1,
				rs2:   rs2,
				label: label,
			})
		case "bne":
			if err := validateArgs(3, elements, remainingLine); err != nil {
				return Application{}, fmt.Errorf("line %s: %v", remainingLine, err)
			}
			rs1, err := parseRegister(strings.TrimSpace(elements[0]))
			if err != nil {
				return Application{}, fmt.Errorf("line %s: %v", remainingLine, err)
			}
			rs2, err := parseRegister(strings.TrimSpace(elements[1]))
			if err != nil {
				return Application{}, fmt.Errorf("line %s: %v", remainingLine, err)
			}
			label := strings.TrimSpace(elements[2])
			instructions = append(instructions, &bne{
				rs1:   rs1,
				rs2:   rs2,
				label: label,
			})
		case "div":
			if err := validateArgs(3, elements, remainingLine); err != nil {
				return Application{}, fmt.Errorf("line %s: %v", remainingLine, err)
			}
			rd, err := parseRegister(strings.TrimSpace(elements[0]))
			if err != nil {
				return Application{}, fmt.Errorf("line %s: %v", remainingLine, err)
			}
			rs1, err := parseRegister(strings.TrimSpace(elements[1]))
			if err != nil {
				return Application{}, fmt.Errorf("line %s: %v", remainingLine, err)
			}
			rs2, err := parseRegister(strings.TrimSpace(elements[2]))
			if err != nil {
				return Application{}, fmt.Errorf("line %s: %v", remainingLine, err)
			}
			instructions = append(instructions, &div{
				rd:  rd,
				rs1: rs1,
				rs2: rs2,
			})
		case "j":
			if err := validateArgs(1, elements, remainingLine); err != nil {
				return Application{}, fmt.Errorf("line %s: %v", remainingLine, err)
			}
			label := strings.TrimSpace(elements[0])
			instructions = append(instructions, &j{
				label: label,
			})
		case "jal":
			if err := validateArgs(2, elements, remainingLine); err != nil {
				return Application{}, fmt.Errorf("line %s: %v", remainingLine, err)
			}
			rd, err := parseRegister(strings.TrimSpace(elements[0]))
			if err != nil {
				return Application{}, fmt.Errorf("line %s: %v", remainingLine, err)
			}
			label := strings.TrimSpace(elements[1])
			instructions = append(instructions, &jal{
				label: label,
				rd:    rd,
			})
		case "jalr":
			if err := validateArgs(3, elements, remainingLine); err != nil {
				return Application{}, fmt.Errorf("line %s: %v", remainingLine, err)
			}
			rd, err := parseRegister(strings.TrimSpace(elements[0]))
			if err != nil {
				return Application{}, fmt.Errorf("line %s: %v", remainingLine, err)
			}
			rs, err := parseRegister(strings.TrimSpace(elements[1]))
			if err != nil {
				return Application{}, fmt.Errorf("line %s: %v", remainingLine, err)
			}
			imm, err := strconv.ParseInt(strings.TrimSpace(elements[2]), 10, 32)
			if err != nil {
				return Application{}, fmt.Errorf("line %s: %v", remainingLine, err)
			}
			instructions = append(instructions, &jalr{
				rd:  rd,
				rs:  rs,
				imm: int32(imm),
			})
		case "lui":
			if err := validateArgs(2, elements, remainingLine); err != nil {
				return Application{}, fmt.Errorf("line %s: %v", remainingLine, err)
			}
			rd, err := parseRegister(strings.TrimSpace(elements[0]))
			if err != nil {
				return Application{}, fmt.Errorf("line %s: %v", remainingLine, err)
			}
			imm, err := strconv.ParseInt(strings.TrimSpace(elements[1]), 10, 32)
			if err != nil {
				return Application{}, fmt.Errorf("line %s: %v", remainingLine, err)
			}
			instructions = append(instructions, &lui{
				rd:  rd,
				imm: int32(imm),
			})
		case "lb":
			if err := validateArgs(2, elements, remainingLine); err != nil {
				return Application{}, fmt.Errorf("line %s: %v", remainingLine, err)
			}
			rd, err := parseRegister(strings.TrimSpace(elements[0]))
			if err != nil {
				return Application{}, fmt.Errorf("line %s: %v", remainingLine, err)
			}
			offset, rs, err := parseOffsetReg(strings.TrimSpace(elements[1]))
			if err != nil {
				return Application{}, fmt.Errorf("line %s: %v", remainingLine, err)
			}
			instructions = append(instructions, &lb{
				rd:     rd,
				offset: offset,
				rs:     rs,
			})
		case "lh":
			if err := validateArgs(2, elements, remainingLine); err != nil {
				return Application{}, fmt.Errorf("line %s: %v", remainingLine, err)
			}
			rd, err := parseRegister(strings.TrimSpace(elements[0]))
			if err != nil {
				return Application{}, fmt.Errorf("line %s: %v", remainingLine, err)
			}
			offset, rs, err := parseOffsetReg(strings.TrimSpace(elements[1]))
			if err != nil {
				return Application{}, fmt.Errorf("line %s: %v", remainingLine, err)
			}
			instructions = append(instructions, &lh{
				rd:     rd,
				offset: offset,
				rs:     rs,
			})
		case "li":
			if err := validateArgs(2, elements, remainingLine); err != nil {
				return Application{}, fmt.Errorf("line %s: %v", remainingLine, err)
			}
			rd, err := parseRegister(strings.TrimSpace(elements[0]))
			if err != nil {
				return Application{}, fmt.Errorf("line %s: %v", remainingLine, err)
			}
			imm, err := strconv.ParseInt(strings.TrimSpace(elements[1]), 10, 32)
			if err != nil {
				return Application{}, fmt.Errorf("line %s: %v", remainingLine, err)
			}
			instructions = append(instructions, &li{
				rd:  rd,
				imm: int32(imm),
			})
		case "lw":
			if err := validateArgs(2, elements, remainingLine); err != nil {
				return Application{}, fmt.Errorf("line %s: %v", remainingLine, err)
			}
			rd, err := parseRegister(strings.TrimSpace(elements[0]))
			if err != nil {
				return Application{}, fmt.Errorf("line %s: %v", remainingLine, err)
			}
			offset, rs, err := parseOffsetReg(strings.TrimSpace(elements[1]))
			if err != nil {
				return Application{}, fmt.Errorf("line %s: %v", remainingLine, err)
			}
			instructions = append(instructions, &lw{
				rd:     rd,
				offset: offset,
				rs:     rs,
			})
		case "nop":
			instructions = append(instructions, &nop{})
		case "mul":
			if err := validateArgs(3, elements, remainingLine); err != nil {
				return Application{}, fmt.Errorf("line %s: %v", remainingLine, err)
			}
			rd, err := parseRegister(strings.TrimSpace(elements[0]))
			if err != nil {
				return Application{}, fmt.Errorf("line %s: %v", remainingLine, err)
			}
			rs1, err := parseRegister(strings.TrimSpace(elements[1]))
			if err != nil {
				return Application{}, fmt.Errorf("line %s: %v", remainingLine, err)
			}
			rs2, err := parseRegister(strings.TrimSpace(elements[2]))
			if err != nil {
				return Application{}, fmt.Errorf("line %s: %v", remainingLine, err)
			}
			instructions = append(instructions, &mul{
				rd:  rd,
				rs1: rs1,
				rs2: rs2,
			})
		case "mv":
			if err := validateArgs(2, elements, remainingLine); err != nil {
				return Application{}, fmt.Errorf("line %s: %v", remainingLine, err)
			}
			rd, err := parseRegister(strings.TrimSpace(elements[0]))
			if err != nil {
				return Application{}, fmt.Errorf("line %s: %v", remainingLine, err)
			}
			rs, err := parseRegister(strings.TrimSpace(elements[1]))
			if err != nil {
				return Application{}, fmt.Errorf("line %s: %v", remainingLine, err)
			}
			instructions = append(instructions, &mv{
				rd: rd,
				rs: rs,
			})
		case "or":
			if err := validateArgs(3, elements, remainingLine); err != nil {
				return Application{}, fmt.Errorf("line %s: %v", remainingLine, err)
			}
			rd, err := parseRegister(strings.TrimSpace(elements[0]))
			if err != nil {
				return Application{}, fmt.Errorf("line %s: %v", remainingLine, err)
			}
			rs1, err := parseRegister(strings.TrimSpace(elements[1]))
			if err != nil {
				return Application{}, fmt.Errorf("line %s: %v", remainingLine, err)
			}
			rs2, err := parseRegister(strings.TrimSpace(elements[2]))
			if err != nil {
				return Application{}, fmt.Errorf("line %s: %v", remainingLine, err)
			}
			instructions = append(instructions, &or{
				rd:  rd,
				rs1: rs1,
				rs2: rs2,
			})
		case "ori":
			if err := validateArgs(3, elements, remainingLine); err != nil {
				return Application{}, fmt.Errorf("line %s: %v", remainingLine, err)
			}
			rd, err := parseRegister(strings.TrimSpace(elements[0]))
			if err != nil {
				return Application{}, fmt.Errorf("line %s: %v", remainingLine, err)
			}
			rs, err := parseRegister(strings.TrimSpace(elements[1]))
			if err != nil {
				return Application{}, fmt.Errorf("line %s: %v", remainingLine, err)
			}
			imm, err := strconv.ParseInt(strings.TrimSpace(elements[2]), 10, 32)
			if err != nil {
				return Application{}, fmt.Errorf("line %s: %v", remainingLine, err)
			}
			instructions = append(instructions, &ori{
				imm: int32(imm),
				rd:  rd,
				rs:  rs,
			})
		case "rem":
			if err := validateArgs(3, elements, remainingLine); err != nil {
				return Application{}, fmt.Errorf("line %s: %v", remainingLine, err)
			}
			rd, err := parseRegister(strings.TrimSpace(elements[0]))
			if err != nil {
				return Application{}, fmt.Errorf("line %s: %v", remainingLine, err)
			}
			rs1, err := parseRegister(strings.TrimSpace(elements[1]))
			if err != nil {
				return Application{}, fmt.Errorf("line %s: %v", remainingLine, err)
			}
			rs2, err := parseRegister(strings.TrimSpace(elements[2]))
			if err != nil {
				return Application{}, fmt.Errorf("line %s: %v", remainingLine, err)
			}
			instructions = append(instructions, &rem{
				rd:  rd,
				rs1: rs1,
				rs2: rs2,
			})
		case "ret":
			instructions = append(instructions, &ret{})
		case "sb":
			if err := validateArgs(2, elements, remainingLine); err != nil {
				return Application{}, fmt.Errorf("line %s: %v", remainingLine, err)
			}
			rs2, err := parseRegister(strings.TrimSpace(elements[0]))
			if err != nil {
				return Application{}, fmt.Errorf("line %s: %v", remainingLine, err)
			}
			offset, rs1, err := parseOffsetReg(strings.TrimSpace(elements[1]))
			if err != nil {
				return Application{}, fmt.Errorf("line %s: %v", remainingLine, err)
			}
			instructions = append(instructions, &sb{
				rs2:    rs2,
				offset: offset,
				rs1:    rs1,
			})
		case "sh":
			if err := validateArgs(3, elements, remainingLine); err != nil {
				return Application{}, fmt.Errorf("line %s: %v", remainingLine, err)
			}
			rs2, err := parseRegister(strings.TrimSpace(elements[0]))
			if err != nil {
				return Application{}, fmt.Errorf("line %s: %v", remainingLine, err)
			}
			offset, err := strconv.ParseInt(strings.TrimSpace(elements[1]), 10, 32)
			if err != nil {
				return Application{}, fmt.Errorf("line %s: %v", remainingLine, err)
			}
			rs1, err := parseRegister(strings.TrimSpace(elements[2]))
			if err != nil {
				return Application{}, fmt.Errorf("line %s: %v", remainingLine, err)
			}
			instructions = append(instructions, &sh{
				rs2:    rs2,
				offset: int32(offset),
				rs1:    rs1,
			})

		case "sll":
			if err := validateArgs(3, elements, remainingLine); err != nil {
				return Application{}, fmt.Errorf("line %s: %v", remainingLine, err)
			}
			rd, err := parseRegister(strings.TrimSpace(elements[0]))
			if err != nil {
				return Application{}, fmt.Errorf("line %s: %v", remainingLine, err)
			}
			rs1, err := parseRegister(strings.TrimSpace(elements[1]))
			if err != nil {
				return Application{}, fmt.Errorf("line %s: %v", remainingLine, err)
			}
			rs2, err := parseRegister(strings.TrimSpace(elements[2]))
			if err != nil {
				return Application{}, fmt.Errorf("line %s: %v", remainingLine, err)
			}
			instructions = append(instructions, &sll{
				rd:  rd,
				rs1: rs1,
				rs2: rs2,
			})

		case "slli":
			if err := validateArgs(3, elements, remainingLine); err != nil {
				return Application{}, fmt.Errorf("line %s: %v", remainingLine, err)
			}
			rd, err := parseRegister(strings.TrimSpace(elements[0]))
			if err != nil {
				return Application{}, fmt.Errorf("line %s: %v", remainingLine, err)
			}
			rs, err := parseRegister(strings.TrimSpace(elements[1]))
			if err != nil {
				return Application{}, fmt.Errorf("line %s: %v", remainingLine, err)
			}
			imm, err := strconv.ParseInt(strings.TrimSpace(elements[2]), 10, 32)
			if err != nil {
				return Application{}, fmt.Errorf("line %s: %v", remainingLine, err)
			}
			instructions = append(instructions, &slli{
				rd:  rd,
				rs:  rs,
				imm: int32(imm),
			})

		case "slt":
			if err := validateArgs(3, elements, remainingLine); err != nil {
				return Application{}, fmt.Errorf("line %s: %v", remainingLine, err)
			}
			rd, err := parseRegister(strings.TrimSpace(elements[0]))
			if err != nil {
				return Application{}, fmt.Errorf("line %s: %v", remainingLine, err)
			}
			rs1, err := parseRegister(strings.TrimSpace(elements[1]))
			if err != nil {
				return Application{}, fmt.Errorf("line %s: %v", remainingLine, err)
			}
			rs2, err := parseRegister(strings.TrimSpace(elements[2]))
			if err != nil {
				return Application{}, fmt.Errorf("line %s: %v", remainingLine, err)
			}
			instructions = append(instructions, &slt{
				rd:  rd,
				rs1: rs1,
				rs2: rs2,
			})

		case "sltu":
			if err := validateArgs(3, elements, remainingLine); err != nil {
				return Application{}, fmt.Errorf("line %s: %v", remainingLine, err)
			}
			rd, err := parseRegister(strings.TrimSpace(elements[0]))
			if err != nil {
				return Application{}, fmt.Errorf("line %s: %v", remainingLine, err)
			}
			rs1, err := parseRegister(strings.TrimSpace(elements[1]))
			if err != nil {
				return Application{}, fmt.Errorf("line %s: %v", remainingLine, err)
			}
			rs2, err := parseRegister(strings.TrimSpace(elements[2]))
			if err != nil {
				return Application{}, fmt.Errorf("line %s: %v", remainingLine, err)
			}
			instructions = append(instructions, &sltu{
				rd:  rd,
				rs1: rs1,
				rs2: rs2,
			})
		case "slti":
			if err := validateArgs(3, elements, remainingLine); err != nil {
				return Application{}, fmt.Errorf("line %s: %v", remainingLine, err)
			}
			rd, err := parseRegister(strings.TrimSpace(elements[0]))
			if err != nil {
				return Application{}, fmt.Errorf("line %s: %v", remainingLine, err)
			}
			rs, err := parseRegister(strings.TrimSpace(elements[1]))
			if err != nil {
				return Application{}, fmt.Errorf("line %s: %v", remainingLine, err)
			}
			imm, err := strconv.ParseInt(strings.TrimSpace(elements[2]), 10, 32)
			if err != nil {
				return Application{}, fmt.Errorf("line %s: %v", remainingLine, err)
			}
			instructions = append(instructions, &slti{
				rd:  rd,
				rs:  rs,
				imm: int32(imm),
			})
		case "sra":
			if err := validateArgs(3, elements, remainingLine); err != nil {
				return Application{}, fmt.Errorf("line %s: %v", remainingLine, err)
			}
			rd, err := parseRegister(strings.TrimSpace(elements[0]))
			if err != nil {
				return Application{}, fmt.Errorf("line %s: %v", remainingLine, err)
			}
			rs1, err := parseRegister(strings.TrimSpace(elements[1]))
			if err != nil {
				return Application{}, fmt.Errorf("line %s: %v", remainingLine, err)
			}
			rs2, err := parseRegister(strings.TrimSpace(elements[2]))
			if err != nil {
				return Application{}, fmt.Errorf("line %s: %v", remainingLine, err)
			}
			instructions = append(instructions, &sra{
				rd:  rd,
				rs1: rs1,
				rs2: rs2,
			})
		case "srai":
			if err := validateArgs(3, elements, remainingLine); err != nil {
				return Application{}, fmt.Errorf("line %s: %v", remainingLine, err)
			}
			rd, err := parseRegister(strings.TrimSpace(elements[0]))
			if err != nil {
				return Application{}, fmt.Errorf("line %s: %v", remainingLine, err)
			}
			rs, err := parseRegister(strings.TrimSpace(elements[1]))
			if err != nil {
				return Application{}, fmt.Errorf("line %s: %v", remainingLine, err)
			}
			imm, err := strconv.ParseInt(strings.TrimSpace(elements[2]), 10, 32)
			if err != nil {
				return Application{}, fmt.Errorf("line %s: %v", remainingLine, err)
			}
			instructions = append(instructions, &srai{
				rd:  rd,
				rs:  rs,
				imm: int32(imm),
			})
		case "srl":
			if err := validateArgs(3, elements, remainingLine); err != nil {
				return Application{}, fmt.Errorf("line %s: %v", remainingLine, err)
			}
			rd, err := parseRegister(strings.TrimSpace(elements[0]))
			if err != nil {
				return Application{}, fmt.Errorf("line %s: %v", remainingLine, err)
			}
			rs1, err := parseRegister(strings.TrimSpace(elements[1]))
			if err != nil {
				return Application{}, fmt.Errorf("line %s: %v", remainingLine, err)
			}
			rs2, err := parseRegister(strings.TrimSpace(elements[2]))
			if err != nil {
				return Application{}, fmt.Errorf("line %s: %v", remainingLine, err)
			}
			instructions = append(instructions, &srl{
				rd:  rd,
				rs1: rs1,
				rs2: rs2,
			})
		case "srli":
			if err := validateArgs(3, elements, remainingLine); err != nil {
				return Application{}, fmt.Errorf("line %s: %v", remainingLine, err)
			}
			rd, err := parseRegister(strings.TrimSpace(elements[0]))
			if err != nil {
				return Application{}, fmt.Errorf("line %s: %v", remainingLine, err)
			}
			rs, err := parseRegister(strings.TrimSpace(elements[1]))
			if err != nil {
				return Application{}, fmt.Errorf("line %s: %v", remainingLine, err)
			}
			imm, err := strconv.ParseInt(strings.TrimSpace(elements[2]), 10, 32)
			if err != nil {
				return Application{}, fmt.Errorf("line %s: %v", remainingLine, err)
			}
			instructions = append(instructions, &srli{
				rd:  rd,
				rs:  rs,
				imm: int32(imm),
			})
		case "sub":
			if err := validateArgs(3, elements, remainingLine); err != nil {
				return Application{}, fmt.Errorf("line %s: %v", remainingLine, err)
			}
			rd, err := parseRegister(strings.TrimSpace(elements[0]))
			if err != nil {
				return Application{}, fmt.Errorf("line %s: %v", remainingLine, err)
			}
			rs1, err := parseRegister(strings.TrimSpace(elements[1]))
			if err != nil {
				return Application{}, fmt.Errorf("line %s: %v", remainingLine, err)
			}
			rs2, err := parseRegister(strings.TrimSpace(elements[2]))
			if err != nil {
				return Application{}, fmt.Errorf("line %s: %v", remainingLine, err)
			}
			instructions = append(instructions, &sub{
				rd:  rd,
				rs1: rs1,
				rs2: rs2,
			})

		case "sw":
			if err := validateArgs(2, elements, remainingLine); err != nil {
				return Application{}, fmt.Errorf("line %s: %v", remainingLine, err)
			}
			rs2, err := parseRegister(strings.TrimSpace(elements[0]))
			if err != nil {
				return Application{}, fmt.Errorf("line %s: %v", remainingLine, err)
			}
			offset, rs1, err := parseOffsetReg(strings.TrimSpace(elements[1]))
			if err != nil {
				return Application{}, fmt.Errorf("line %s: %v", remainingLine, err)
			}
			instructions = append(instructions, &sw{
				rs2:    rs2,
				offset: offset,
				rs1:    rs1,
			})
		case "xor":
			if err := validateArgs(3, elements, remainingLine); err != nil {
				return Application{}, fmt.Errorf("line %s: %v", remainingLine, err)
			}
			rd, err := parseRegister(strings.TrimSpace(elements[0]))
			if err != nil {
				return Application{}, fmt.Errorf("line %s: %v", remainingLine, err)
			}
			rs1, err := parseRegister(strings.TrimSpace(elements[1]))
			if err != nil {
				return Application{}, fmt.Errorf("line %s: %v", remainingLine, err)
			}
			rs2, err := parseRegister(strings.TrimSpace(elements[2]))
			if err != nil {
				return Application{}, fmt.Errorf("line %s: %v", remainingLine, err)
			}
			instructions = append(instructions, &xor{
				rd:  rd,
				rs1: rs1,
				rs2: rs2,
			})

		case "xori":
			if err := validateArgs(3, elements, remainingLine); err != nil {
				return Application{}, fmt.Errorf("line %s: %v", remainingLine, err)
			}
			rd, err := parseRegister(strings.TrimSpace(elements[0]))
			if err != nil {
				return Application{}, fmt.Errorf("line %s: %v", remainingLine, err)
			}
			rs, err := parseRegister(strings.TrimSpace(elements[1]))
			if err != nil {
				return Application{}, fmt.Errorf("line %s: %v", remainingLine, err)
			}
			imm, err := strconv.ParseInt(strings.TrimSpace(elements[2]), 10, 32)
			if err != nil {
				return Application{}, fmt.Errorf("line %s: %v", remainingLine, err)
			}
			instructions = append(instructions, &xori{
				imm: int32(imm),
				rd:  rd,
				rs:  rs,
			})
		default:
			return Application{}, fmt.Errorf("invalid instruction type: %s", line)
		}
		pc += 4
	}

	return Application{
		Instructions: instructions,
		Labels:       labels,
	}, nil
}

func validateArgs(expected int, args []string, line string) error {
	if len(args) != expected {
		return fmt.Errorf("invalid line: expected %d arguments, got %d: %v", expected, len(args), line)
	}
	return nil
}

func validateArgsInterval(min, max int, args []string, line string) error {
	if len(args) >= min && len(args) <= max {
		return nil
	}
	return fmt.Errorf("invalid line: expected between %d and %d arguments, got %d: %v", min, max, len(args), line)
}

func parseRegister(s string) (RegisterType, error) {
	switch s {
	case "zero", "$zero":
		return Zero, nil
	case "ra", "$ra":
		return Ra, nil
	case "sp", "$sp":
		return Sp, nil
	case "gp", "$gp":
		return Gp, nil
	case "tp", "$tp":
		return Tp, nil
	case "t0", "$t0":
		return T0, nil
	case "t1", "$t1":
		return T1, nil
	case "t2", "$t2":
		return T2, nil
	case "s0", "$s0":
		return S0, nil
	case "s1", "$s1":
		return S1, nil
	case "a0", "$a0":
		return A0, nil
	case "a1", "$a1":
		return A1, nil
	case "a2", "$a2":
		return A2, nil
	case "a3", "$a3":
		return A3, nil
	case "a4", "$a4":
		return A4, nil
	case "a5", "$a5":
		return A5, nil
	case "a6", "$a6":
		return A6, nil
	case "a7", "$a7":
		return A7, nil
	case "s2", "$s2":
		return S2, nil
	case "s3", "$s3":
		return S3, nil
	case "s4", "$s4":
		return S4, nil
	case "s5", "$s5":
		return S5, nil
	case "s6", "$s6":
		return S6, nil
	case "s7", "$s7":
		return S7, nil
	case "s8", "$s8":
		return S8, nil
	case "s9", "$s9":
		return S9, nil
	case "s10", "$s10":
		return S10, nil
	case "s11", "$s11":
		return S11, nil
	case "t3", "$t3":
		return T3, nil
	case "t4", "$t4":
		return T4, nil
	case "t5", "$t5":
		return T5, nil
	case "t6", "$t6":
		return T6, nil
	default:
		return 0, fmt.Errorf("unknown register: %v", s)
	}
}

func parseOffsetReg(s string) (int32, RegisterType, error) {
	firstParenthesis := strings.IndexRune(s, '(')
	if firstParenthesis == -1 {
		return 0, 0, fmt.Errorf("invalid offset register: %s", s)
	}

	immString := strings.TrimSpace(s[:firstParenthesis])
	imm, err := strconv.ParseInt(immString, 10, 32)
	if err != nil {
		return 0, 0, err
	}

	regString := strings.TrimSpace(s[firstParenthesis+1 : len(s)-1])

	reg, err := parseRegister(regString)
	if err != nil {
		return 0, 0, err
	}

	return int32(imm), reg, nil
}
