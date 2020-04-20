source = ["./terrafmt"]
bundle_id = "com.terrycain.github.terrafmt"

apple_id {
  username = "terry@dolphincorp.co.uk"
  password = "@env:APPLE_APP_PW"
}

sign {
  application_identity = "Developer ID Application: Terry Cain (UT7M7Z36B6)"
}

#dmg {
#  output_path = "terrafmt.dmg"
#  volume_name = "Terraform Formatting Tool"
#}

zip {
  output_path = "terrafmt.zip"
}
