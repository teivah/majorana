    # Init by storing value to memory
    addi t0, zero, %d
    sw t0, 0, zero

    addi t0, zero, 0 # Address of the word
    lw t0, 0(t0) # Load word in memory

    # Compute max
    addi t1, zero, 2 # 4 t1 = 2
    div t1, t0, t1   # 5 t1 = num / t1 (2)
    addi t1, t1, 1   # 6 t1++

    addi t2, zero, 2 # Counter init

loop:
    bge t2, t1, true # 8 While loop
    rem t3, t0, t2 # Modulo
    beq t3, zero, false # If equals 0
    addi t2, t2, 1 # Increment counter
    jal zero, loop # 12

true:
    addi t0, zero, 1 # 13
    jal zero, end

false:
    addi t0, zero, 0 # 15
    jal zero, end

end:
    addi t1, zero, 4 # 17
    sb t0, 0, t1 # Stores t0 into memory t1 + 0 = 4 + 0
    addi a0, t1, 0
    addi ra, zero, 0