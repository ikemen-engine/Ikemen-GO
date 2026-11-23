# Ordered intro phases

This branch adds two `AssertSpecial` flags to the existing `intro` flag:

beforeIntro: Executes before the standard intro phase.
intro: The existing standard intro phase.
afterIntro: Executes after the standard intro phase.

These phases are not specific to characters or stage elements.
Both characters and stage elements can use AssertSpecial
to participate in any of the three phases.

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


The scheduler is enabled only when `beforeIntro` or `afterIntro` is actually asserted. A project that only uses the original `intro` flag follows the previous behavior.

When the extended scheduler is active, pressing a button during any active intro phase starts the existing shutter skip effect and freezes all actors in the intro scheduler. When the shutter signals completion, the engine skips the remaining intro phases and resumes at the normal round-announcement point. The regular Round and Fight announcements are then displayed before `RoundState = 2` begins.

The flag names are case-insensitive in CNS, so `beforeIntro`, `beforeintro`, and `BEFOREINTRO` are accepted.
