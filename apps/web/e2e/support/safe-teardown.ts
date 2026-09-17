/**
 * Runs a cleanup step and swallows any error it throws.
 *
 * Decisions.md §J7: fixture teardown runs after an assertion failure *and*
 * after a test timeout, so a failed cleanup must never become the reported
 * failure — it would mask the real one. Errors are logged instead so a
 * leftover `E2E_`-prefixed row stays findable (see §J8's orphan query)
 * without ever hiding what the test actually failed on.
 *
 * This is also what makes `CategoryFactory.cleanup()` safe when a category
 * still has a tour attached (its delete button is disabled by the UI, so
 * the click/confirm sequence times out rather than succeeding): the error
 * is caught here, logged, and cleanup moves on to the next created entity.
 */
export async function safeTeardown(what: string, fn: () => Promise<void>): Promise<void> {
  try {
    await fn();
  } catch (error) {
    console.warn(`[e2e] cleanup failed for ${what}: ${(error as Error).message}`);
  }
}
