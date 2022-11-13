module.exports = {
  valid: (val) => {
    if (/(?:^v([0-9]+).([0-9]+).([0-9]+))(-([a-z]+)(.[0-9])?)?$/i.test(val)) {
      return true;
    }

    throw new Error(`Value ${val} is not a valid semver version`)
  }
}