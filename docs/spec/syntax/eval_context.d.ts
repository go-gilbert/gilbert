/**
 * @schema
 * This file describes a schema of available variables in Expr lang expressions using TypeScript type syntax.
 * It is NOT part of the application codebase (which is in Go).
 * Do not generate TypeScript code from this file.
 *
 * See:
 *  - @./schema.d.ts - workflow file (`gilbert.yaml`) schema.
 *  - @./expressions.md - expression language documentation.
 */

/**
 * Holds global scope available in `${{...}}` expressions.
 *
 * Acts like `globalThis`/`window` in JavaScript.
 */
interface Scope {
  /**
   * Carries values provided to task/mixin inputs.
   *
   * @see [JobBase.on] in workflow file schema.
   */
  inputs?: Record<string, any>

  /**
   * Holds constants defined in `const` section of workflow file.
   *
   * @see [WorkflowFile.const]
   */
  consts?: Record<string, any>

  project: {
    /**
     * Current job working directory.
     */
    workDir: string

    /**
     * Directory where workflow file is located.
     */
    workspaceDir: string

    /**
     * Path to workflow file.
     */
    workflowFile: string
  }

  /**
   * System environment variables.
   */
  env: Record<string, string>

  /**
   * Matrix values populated to a matrix strategy job.
   *
   * Only available if job has `.strategy.matrix` defined.
   *
    * @see [JobBase.strategy.matrix]
    */
  matrix?: Record<string, any>

  /**
   * Metadata provided with a signal, fired by action.
   *
   * Only available if job is defined in `.on` block.
   *
   * @see [JobBase.on]
   */
  event?: Record<string, any>
}
