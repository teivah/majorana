main:
  li t2, 2       # 0
  lw t0, 0(zero) # 1
  beqz t0, 2     # 2
  li t1, 1       # 3
1:
  ret            # 4
2:
  ret            # 5