require 'json'
require 'net/http'
require 'uri'
require 'openssl'
old_host = "http://sitrep-vatcinc.com"
new_host = "https://sitrep-demo.herokuapp.com"

puts "Migrating Intellipedia"
intellipedia_pages = JSON.parse(`http #{old_host}/api/v2/intellipedia`)

uri = URI.parse(new_host)
http = Net::HTTP.new(uri.host, uri.port)
http.use_ssl = true
http.verify_mode = OpenSSL::SSL::VERIFY_NONE
request = Net::HTTP::Post.new("/api/v2/intellipedia")
request.add_field('Content-Type', 'application/json')


intellipedia_pages.each do |page|
  #puts page.to_json
  request.body = {intel: page}.to_json
  response = http.request(request)
  puts response.body
end

# puts "Migrating News"
#
# puts "Migrating files"
# images = JSON.parse(`http #{old_host}/api/v2/settings`)["images"]
# images.each do |title, img|
#   path = img.split("/").last
#   if img.include?("http")
#     `curl -sSL #{img} -o #{path}`
#   else
#     `curl -sSL #{old_host}#{img} -o #{path}`
#   end
#
#   new_res = JSON.parse(`http --form POST #{new_host}/api/v2/upload content@#{path}`)
#   `rm #{path}`
#   `http PUT #{new_host}/api/v2/img/#{title}?vValue=#{new_res["url"]}`
# end
#img_uploaded =
