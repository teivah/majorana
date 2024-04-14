bubsort:
    # a0 = long *list
    # a1 = size
    # t0 = swapped
    # t1 = i
1:
    li t0, 0          # swapped = false
    li t1, 1          # i = 1
2:
    bge t1, a1, 4     # break if i >= size
    slli t3, t1, 2    # scale i by 4
    add t3, a0, t3    # new scaled memory address
    lw  t4, -4(t3)    # load list[i-1] into t4
    lw  t5, 0(t3)     # load list[i] into t5
    ble t4, t5, 3     # if list[i-1] < list[i], it's in position
    # if we get here, we need to swap
    li  t0, 1         # swapped = true
    sw  t4, 0(t3)     # list[i] = list[i-1]
    sw  t5, -4(t3)    # list[i-1] = list[i]
3:
    addi t1, t1, 1    # i++
    j    2            # loop again
4:
    bnez t0, 1        # loop if swapped = true
    ret               # return via return address register