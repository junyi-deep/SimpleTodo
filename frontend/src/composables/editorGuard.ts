// Exactly one inline task editor can be active in the workspace.
let active: { dirty: () => boolean; leave: () => Promise<boolean> } | undefined;
export function registerEditorGuard(
  dirty: () => boolean,
  leave: () => Promise<boolean>,
) {
  const owner = { dirty, leave };
  active = owner;
  return () => {
    if (active === owner) active = undefined;
  };
}
export const hasDirtyEditor = () => active?.dirty() || false;
export const canLeaveEditor = async () => (active ? active.leave() : true);
