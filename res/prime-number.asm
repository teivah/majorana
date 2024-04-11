    # Init by storing value to memory
    lw t0, 0(zero)   # 0 t0 = %d

    # Compute max
    addi t1, zero, 2 # 1 t1 = 2
    div t1, t0, t1   # 2 t1 = t0 / t1
    addi t1, t1, 1   # 3 t1++

    addi t2, zero, 2 # 4 t2 = 2

loop:
    bge t2, t1, true # 5 While loop
    rem t3, t0, t2 # Modulo
    beq t3, zero, false # If equals 0
    addi t2, t2, 1 # Increment counter
    j loop

true:
    addi t0, zero, 1 # 10
    j end

false:
    addi t0, zero, 0 # 12
    j end

end:
    addi t1, zero, 4 # 14
    sb t0, 0(t1) # Stores t0 into memory t1 + 0 = 4 + 0
    addi a0, t1, 0
    addi ra, zero, 0