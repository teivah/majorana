# 3[1,2,3]______9
# -
# Array size
#   ------      -
#   Visible   Invisible
#
main:
  lw t0, 0(zero)   # t0 represents the array size
  li t1, 1         # t1 represents the array
  li t2, 9         # t2 represents the secret
  bge t2, t0, end
  sw t0, 4(t1)
end:
  ret              # Value in memory[4]?