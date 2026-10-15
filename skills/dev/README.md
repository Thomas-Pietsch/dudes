# Dev-level skills

For the dev repos: code and tests. Models are trained on code, so a worker here needs few skills (`CORE.md` §5, *Where skills earn their place*): the goals and tests carry the work, the model gets as much freedom as possible, and DUDES improves as models do.

Every skill here:

- is opt-in for a particular task — a test-first slice, a review, a hard bug, a refactoring — named by a recipe or contract task (F48), never given by default (F76);
- stays short: the engine inlines it into a worker's prompt, under the context budget;
- names its upstream source and revision in `metadata` when it adapts one, so a refresh is a diff, not a rewrite.

Which role loads which skill in the role-play: `rules/DUDES.md` §1.
