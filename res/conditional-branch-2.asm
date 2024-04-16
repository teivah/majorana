main:
  li t2, 2
  lw t0, 0(zero)
  beqz t0, 2
  li t1, 1
1:
  ret
2:
  ret