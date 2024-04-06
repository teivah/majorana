strncpy:
    # a0 = char *dst
    # a1 = const char *src
    # a2 = unsigned long n
    # t0 = i
    li      t0, 0        # 0 i = 0
1:
    bge     t0, a2, 2    # 1 break if i >= n
    add     t1, a1, t0   # 2 src + i
    lb      t1, 0(t1)    # 3 t1 = src[i]
    beqz    t1, 2        # 4 break if src[i] == '\0'
    add     t2, a0, t0   # 5 t2 = dst + i
    sb      t1, 0(t2)    # 6 dst[i] = src[i]
    addi    t0, t0, 1    # 7 i++
    j       1            # 8 back to beginning of loop
2:
    bge     t0, a2, 3    # break if i >= n
    add     t1, a0, t0   # t1 = dst + i
    sb      zero, 0(t1)  # dst[i] = 0
    addi    t0, t0, 1    # i++
    j       2            # back to beginning of loop
3:
    ret                  # return via return address register