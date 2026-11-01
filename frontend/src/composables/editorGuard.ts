// A workspace may have both an inline task editor and an unsaved note.
type Guard = { dirty: () => boolean; leave: () => Promise<boolean> };
const guards = new Set<Guard>();
export function registerEditorGuard(
  dirty: () => boolean,
  leave: () => Promise<boolean>,
) {
  const guard = { dirty, leave };
  guards.add(guard);
  return () => guards.delete(guard);
}
export const hasDirtyEditor = () => [...guards].some((g) => g.dirty());
export const canLeaveEditor = async () => {
  for (const guard of [...guards].reverse())
    if (!(await guard.leave())) return false;
  return true;
};
