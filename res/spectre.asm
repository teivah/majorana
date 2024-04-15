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
  lw t2, 9(zero) # What's the state of t2?
end:
  ret