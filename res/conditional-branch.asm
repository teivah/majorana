main:
  lw t0, 0(zero)
  beqz t0, end
  li t1, 1
end:
  ret