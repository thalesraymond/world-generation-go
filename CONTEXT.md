# Domain Glossary

## faction

The social group a settlement belongs to. Settlements are the only members;
there is no faction-level entity.

A faction is an **absolute loyalty boundary for hostile agent actions**: a
settlement never raids or conquers another settlement of the same faction,
even when relations between them are deeply negative. Same-faction pairs can
hold hostile relations (e.g. −0.8 after a conquest absorbs the target into
the conqueror's faction), but that hostility is never acted on directly — it
only heals through positive drift. Membership changes only by conquest, which
transfers the conquered settlement into the conqueror's faction.

(canonical term resolved in [wayfinder ticket: Prevent same-faction raids in agent action
selection](https://github.com/thalesraymond/world-generation-go/issues/40))