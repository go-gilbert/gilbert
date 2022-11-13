version = 2

module "semver" {
  script "semver" {
    type = "js"
    path = path_join(module.directory, "script.js")
  }

  function "valid" "Validates string as semver version" {
    returns = "boolean"

    // Can also be declared as anonymous "script" block
    script = scripts.semver // also available as "module.scripts.semver"

    // Pass execution context as first result
    passContext = false

    // Will return promise?
    async = false

    param "value" {
      type = "string"
      required = true
      example = "v1.2.3"
    }
  }
}