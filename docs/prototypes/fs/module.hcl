version = 2

module "fs" {
  action "watch" {
    param "path" {
      type = array(string)
    }

    param "ignore" {
      type = array(string)
    }

    hook "onChange" {}
  }
}