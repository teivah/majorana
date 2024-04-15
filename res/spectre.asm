# 3[1,2,3]______9
# -
# Array size
#   ------      -
#   Visible   Invisible
#
main:
  li t1, 9
  lw t0, 0(zero)
  bge t1, t0, end
  lw t2, 36(zero)
end:
  sw t2, 0(zero)
  ret