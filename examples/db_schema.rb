#!/usr/bin/env ruby

src_dir = File.dirname(Dir.getwd + '/' + $0) + '/../'

doc = `GOPATH='#{src_dir}' godoc hammy`

r = Regexp.new('CREATE\s+TABLE.*?ENGINE=InnoDB\s+DEFAULT\s+CHARSET=utf8',
               Regexp::MULTILINE | Regexp::IGNORECASE)

puts 'CREATE DATABASE `hammy`;'
puts 'USE `hammy`;'
puts

while doc do
    r =~ doc
    puts $~.to_s + ';'
    puts
    doc = $'
end
