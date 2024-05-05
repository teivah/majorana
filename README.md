# Majorana

[Majorana](https://en.wikipedia.org/wiki/Ettore_Majorana) is a RISC-V virtual processor written in Go.

## Majorana Virtual Processors (MVP)

### MVP-1

MVP-1 is the first version of the RISC-V virtual machine.
It does not implement any of the known CPU optimizations, such as pipelining, out-of-order execution, multiple execution units, etc.

Here is the microarchitecture, divided into 4 classic stages:
* Fetch: fetch an instruction from the main memory
* Decode: decode the instruction
* Execute: execute the RISC-V instruction
* Write: write-back the result to a register or the main memory

![](res/majorana-mvp-1.drawio.png)

### MVP-2

Compared to MVP-1, we add a cache for instructions called L1I (Level 1 Instructions) with a size of 64 KB. The caching policy is straightforward: as soon as we meet an instruction that is not present in L1I, we fetch a cache line of 64 KB instructions from the main memory, and we cache it into L1I.

![](res/majorana-mvp-2.drawio.png)

### MVP-3

From MVP-3, we introduce a proper memory management unit (MMU). The MMU consists of an 1KB L1I (present from L1I) and a new 1KB L1D to cache data:

![](res/majorana-mvp-3.drawio.png)

When the execute unit wants to access a memory address, it requests it to the MMU that either returns the value directly from L1D or from memory. In the latter case, the MMU fetches a whole cache line of 64 bytes from memory and push that into L1D. L1D eviction policy is based on a LRU cache (Least-Recently Used).

The introduction of an L1D doesn't have any impact for benchmarks not reliant on frequent memory access (obviously); however, it yields significant performance improvements for those that do (up to 40% faster).

### MVP-4

MVP-4 keeps the same microarchitecture as MVP-4 with 4 stages and an MMU. Yet, this version implements [pipelining](https://en.wikipedia.org/wiki/Instruction_pipelining).

In a nutshell, pipelining allows keeping every stage as busy as possible. For example, as soon as the fetch unit has fetched an instruction, it will not wait for the instruction to be decoded, executed and written. It will fetch another instruction straight away during the next cycle(s).

This way, the first instruction can be executed in 4 cycles (assuming the fetch is done from L1I), whereas the next instructions will be executed in only 1 cycle.

One of the complexity with pipelining is to handle branches. What if we fetch a [bge](https://msyksphinz-self.github.io/riscv-isadoc/html/rvi.html#bge) instruction for example? The next instruction fetched will not be necessarily the one we should have fetched/decoded/executed/written. As a solution, we implemented the first version of branch prediction handled by the branch unit.

The branch unit takes the hypothesis that a conditional branch will **not** be taken. Hence, after having fetched an instruction, regardless if it's a conditional branch, we will fetch the next instruction after it. If the prediction was wrong, we need to flush the pipeline, revert the program counter to the destination marked by the conditional branch instruction, and continue the execution.

Pipeline flushing has a significant performance penalty as it requires discarding partially completed instruction and restarting the pipeline, leading to wasted cycles.

There is another problem with pipelining. We might face what we call a data hazard. For example:

```asm
addi t1, zero, 2 # Write to t1
div t2, t0, t1   # Read from t1
``` 

The processor must wait for `ADDI` to be executed and to get its result written in T1 before to execute `DIV` (as div depends on T1).
In this case, we implement what we call pipeline interclock by delaying the execution of `DIV`.

![](res/majorana-mvp-4.drawio.png)

### MVP-5

One issue with MVP-3 is when it met an unconditional branches. For example:

```asm
main:
  jal zero, foo    # Branch to foo
  addi t1, t0, 3   # Set t1 to t0 + 3
foo:
  addi t0, zero, 2 # Set t0 to 2
  ...
```

In this case, the fetch unit, after fetching the first line (`jal`), was fetching the second line (first `addi`), which ended up being a problem because the execution is branching to line 3 (second `addi`). It was resolved by flushing the whole pipeline, which is very costly.

The microarchitecture of MVP-4 is very similar to MVP-3, except that the branch unit is now coupled with a Branch Target Buffer (BTB):

![](res/majorana-mvp-5.drawio.png)

One the fetch unit fetches a branch, it doesn't know whether it's a branch; it's the job of the decode unit. Therefore, the fetch unit can't simply say: "_I fetched a branch, I'm going to wait for the execute unit to tell me the next instruction to fetch_".

The workflow is now the following:
- The fetch unit fetches an instruction.
- The decode unit decodes it. If it's a branch, it waits until the execute unit resolves the destination address.
- When the execute unit resolves the target address of the branch, it notifies the branch unit, with the target address.
- Then, the branch unit notifies the fetch unit, which invalidates the latest instruction fetched.

This helps in preventing a full pipeline flush. Facing an unconditional branch now takes only a few cycles to be resolved.

### MVP-6

#### MVP-6.0

The next step is to implement a so-called superscalar processor. A superscalar processor can execute multiple instructions during a clock cycle by dispatching multiple instructions to different execution units. This is one of the magical things with modern CPUs: even sequential code can be executed in parallel!

The fetch unit and the decode unit are now capable to fetch/decode two instruction within a single cycle. Yet, before to dispatch the executions to the execute units, a new stage comes in: the control unit.

![](res/majorana-mvp-6.drawio.png)

The control unit plays a pivotal role in coordinating the execution of multiple instructions simultaneously. It performs dependency checking between the decoded instructions to guarantee it won't lead to any hazard.

One _small_ issue: MVP-5.0 is not always faster in all the benchmarks. Indeed, when an application is branch-heavy, it performed slightly worst that MVP-4. The main reason being that the control unit logic is very basic and because of that, on average it dispatches less than 0.6 instructions per cycle. Yet, if branches are scarce, it performs significantly better than MVP-4 (~40% in the string copy benchmark).

// TODO: Slower? Because L3

#### MVP-6.1

For MVP-6.1, the microarchitecture is the same as MVP-6.1. The only difference lies in the control unit, where we started to implement a new concept called forwarding. Consider a data hazard mentioned previously:

```asm
addi t1, zero, 2 # Write to t1
div t2, t0, t1   # Read from t1
``` 

Instruction 1 writes to `T1`, while instruction 2 reads from `T2`. Therefore, instruction 2 has to wait for `ADDI` to write the result to `T1` before it gets executed, hence slowing down the execution. With forwarding, we can alleviate the effects of this problem: the result of the `ADDI` instruction is fed directly back into the ALU's input port. `DIV` doesn't have to wait for the execution of `ADDI` to be written in `T1` anymore.

#### MVP-6.2

// TODO MVP-6.1 was vulnerable to branch condition tests initially as the CU allows more instructions

```asm
main:
    lw t0, 0(zero) # memory[0] = 0
    beqz t0, end   # Branch should be taken
    li t1, 1       # Because of speculative instruction this instruction is executed
end:
    ret            # t1 = 1 instead of 0
```

New feature: commit / rollback conditional branches

#### MVP-6.3

Register renaming to tackle waw and war

### MVP-7

#### MVP-7.0

L1D per core, MSI protocol implementation; yet, lots of "false sharing"

#### MVP-7.1

The CU syncs to the MSI when there was an eviction (lost a cycle) but then tells that a particular instruction has to be executed by a specific core.

## Benchmarks

All the benchmarks are executed at a fixed CPU clock frequency of 3.2 GHz.

Meanwhile, we have executed a benchmark on an Apple M1 (same CPU clock frequency). This benchmark was on a different micro-architecture, different ISA, etc. is hardly comparable with the MVP benchmarks. Yet, it gives us a reference to show how good (or bad :) the MVP implementations are.

| Machine | Prime number | Sum of array | String copy | String length | Bubble sort | Avg |
|:------:|:-----:|:-----:|:-----:|:-----:|:-----:|:-----:|
| Apple M1 | 31703.0 ns | 1300.0 ns | 3232.0 ns | 3231.0 ns | 42182.0 ns | 1.0 |
| MVP-1 | 24365508 ns, 768.6x slower | 3252967 ns, 2502.3x slower | 10109189 ns, 3127.8x slower | 6131992 ns, 1897.9x slower | 49641410 ns, 1176.8x slower | 1894.7 |
| MVP-2 | 422921 ns, 13.3x slower | 510824 ns, 392.9x slower | 2275302 ns, 704.0x slower | 1235514 ns, 382.4x slower | 13409097 ns, 317.9x slower | 362.1 |
| MVP-3 | 422922 ns, 13.3x slower | 145409 ns, 111.9x slower | 1313097 ns, 406.3x slower | 273310 ns, 84.6x slower | 1993983 ns, 47.3x slower | 132.7 |
| MVP-4 | 141242 ns, 4.5x slower | 108045 ns, 83.1x slower | 1213745 ns, 375.5x slower | 212257 ns, 65.7x slower | 1495782 ns, 35.5x slower | 112.9 |
| MVP-5 | 125594 ns, 4.0x slower | 106765 ns, 82.1x slower | 1204146 ns, 372.6x slower | 209058 ns, 64.7x slower | 1483345 ns, 35.2x slower | 111.7 |
| MVP-6.0 | 125581 ns, 4.0x slower | 104288 ns, 80.2x slower | 1204999 ns, 372.8x slower | 209992 ns, 65.0x slower | 855390 ns, 20.3x slower | 108.5 |
| MVP-6.1 | 109932 ns, 3.5x slower | 100448 ns, 77.3x slower | 1198406 ns, 370.8x slower | 200391 ns, 62.0x slower | 836670 ns, 19.8x slower | 106.7 |
| MVP-6.2 | 109932 ns, 3.5x slower | 100448 ns, 77.3x slower | 1198406 ns, 370.8x slower | 200391 ns, 62.0x slower | 836670 ns, 19.8x slower | 106.7 |
| MVP-6.3 | 94284 ns, 3.0x slower | 100448 ns, 77.3x slower | 611271 ns, 189.1x slower | 200391 ns, 62.0x slower | 836670 ns, 19.8x slower | 70.2 |
| MVP-7.0 | 94286 ns, 3.0x slower | 42893 ns, 33.0x slower | 94688 ns, 29.3x slower | 51136 ns, 15.8x slower | 7572730 ns, 179.5x slower | 52.1 |
| MVP-7.1 | 94286 ns, 3.0x slower | 42893 ns, 33.0x slower | 94688 ns, 29.3x slower | 51136 ns, 15.8x slower | 384364 ns, 9.1x slower | 18.0 |
| MVP-7.2 | 94332 ns, 3.0x slower | 39623 ns, 30.5x slower | 87494 ns, 27.1x slower | 50218 ns, 15.5x slower | 302969 ns, 7.2x slower | 16.7 |
