strlen:
    # a0 = const char *str
    li     t0, 0         # 0 i = 0
1b:
    add    t1, t0, a0    # 1 Add the byte offset for str[i]
    lb     t1, 0(t1)     # 2 Dereference str[i]
    beqz   t1, 1f        # 3 if str[i] == 0, break for loop
    addi   t0, t0, 1     # 4 Add 1 to our iterator
    j      1b            # 5 Jump back to condition (1 backwards)
1f:
    sw t0, 0(zero)       # 6
    ret                  # 7 Return back via the return address register