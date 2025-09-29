locals {
  zone_tags = merge(var.tags, {
    Name = "${var.zone_name}-zone"
  })
}

resource "aws_route53_zone" "this" {
  count = var.create_zone ? 1 : 0

  name          = var.zone_name
  comment       = var.comment
  force_destroy = var.force_destroy

  tags = local.zone_tags
}

data "aws_route53_zone" "lookup" {
  count = var.create_zone || var.existing_zone_id != null ? 0 : 1

  name         = var.zone_name
  private_zone = false
}

locals {
  zone_id = var.create_zone
    ? aws_route53_zone.this[0].zone_id
    : var.existing_zone_id != null
      ? var.existing_zone_id
      : data.aws_route53_zone.lookup[0].zone_id

  record_map = {
    for idx, rec in var.records :
    "${rec.name}_${upper(rec.type)}_${idx}" => {
      name            = rec.name
      type            = upper(rec.type)
      ttl             = lookup(rec, "ttl", null)
      allow_overwrite = lookup(rec, "allow_overwrite", false)
      values = [for v in lookup(rec, "values", []) : trimspace(v) if v != null && trimspace(v) != ""]
    }
  }

  valid_records = {
    for key, rec in local.record_map :
    key => rec if length(rec.values) > 0
  }
}

resource "aws_route53_record" "this" {
  for_each = local.valid_records

  zone_id = local.zone_id
  name = (
    each.value.name == "" || each.value.name == "@"
  ) ? var.zone_name : (
    endswith(lower(each.value.name), lower(var.zone_name))
    ? each.value.name
    : "${each.value.name}.${var.zone_name}"
  )
  type            = each.value.type
  ttl             = each.value.ttl != null ? each.value.ttl : var.default_ttl
  records         = each.value.values
  allow_overwrite = each.value.allow_overwrite
}
