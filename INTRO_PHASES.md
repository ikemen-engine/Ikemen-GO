# Ordered intro phases

This branch adds two `AssertSpecial` flags to the existing `intro` flag:

- `beforeIntro`: stage-intro phase, before the ordinary character intros.
- `intro`: the existing character-intro phase.
- `afterIntro`: sequence after the ordinary character intros and before the fight begins.

All three phases remain inside `RoundState = 1`. `RoundState = 0` (pre-intro/fade-in) and `RoundState = 2` (fight) keep their existing meanings.

## CNS usage

An actor participating in the first phase should assert the new flag for every frame that its phase state is active:

```ini
[State 190, Stage intro phase]
type = AssertSpecial
triggerall = RoundState = 1
trigger1 = 1
flag = beforeIntro
```

Ordinary character intros continue using the existing flag:

```ini
[State 190, Character intro phase]
type = AssertSpecial
triggerall = RoundState = 1
trigger1 = 1
flag = intro
```

An actor participating in the last phase uses:

```ini
[State 190, After-intro phase]
type = AssertSpecial
triggerall = RoundState = 1
trigger1 = 1
flag = afterIntro
```

A phase ends when all actors asserting its flag leave those states. Multiple actors in the same phase run concurrently. Actors already declaring a later phase are held by the engine until that phase becomes active; their state time, animation, projectiles, and explods do not advance while held. Characters that do not declare one of these flags are not placed in the phase scheduler, preserving the behavior of unrelated existing content. Helpers without their own phase flag inherit a hold from a parent that is waiting for a later phase.

**Every actor that participates in a later phase must declare its phase from the first `RoundState = 1` tick.** For example, an `afterIntro` actor should enter a waiting state that asserts `afterIntro` immediately, and remain in that phase until its turn. Do not wait until the previous phase has ended before asserting the flag: the phase order is monotonic and cannot go backwards.

The engine discovers the actors by evaluating their state controllers on the first `RoundState = 1` tick. Because that first evaluation is required to identify each actor's phase, a later-phase actor can receive one initial tick before the scheduler holds it. From the following tick onward, its state time and animation remain frozen until its phase becomes active.

The scheduler is enabled only when `beforeIntro` or `afterIntro` is actually asserted. A project that only uses the original `intro` flag follows the previous behavior.

## Skip

When the extended scheduler is active, pressing a button during any active intro phase starts the existing shutter skip effect and freezes all actors in the intro scheduler. When the shutter signals completion, the engine skips the remaining intro phases and resumes at the normal round-announcement point. The regular Round and Fight announcements are then displayed before `RoundState = 2` begins.

The flag names are case-insensitive in CNS, so `beforeIntro`, `beforeintro`, and `BEFOREINTRO` are accepted.
