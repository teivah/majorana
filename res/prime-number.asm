    addi t0, zero, 0 # Address of the word
    lw t0, 0, t0 # Load word in memory

    # Compute max
    addi t1, zero, 2
    div t1, t0, t1
    addi t1, t1, 1

    addi t2, zero, 2 # Counter init

loop:
    bge t2, t1, true # While loop
    rem t3, t0, t2 # Modulo
    beq t3, zero, false # If equals 0
    addi t2, t2, 1 # Increment counter
    jal zero, loop

true:
    addi t0, zero, 1
    jal zero, end

false:
    addi t0, zero, 0
    jal zero, end

end:
    addi t1, zero, 4
    sb t0, 0, t1 # Store to address 4
    addi a0, t1, 0