#!/usr/bin/env ruby

require 'msgpack'

class EvalEnv
	def initialize(state, objectname)
		@state = state
		@objectname = objectname
		@cmdbuf = []
	end

	def get_state(key)
		elem = @state[key]
		if elem then
			elem["Value"]
		else
			nil
		end
	end

	def set_state(key, value)
		@state[key] = {
			"Value" => value,
			"LastUpdate" => Time.now.to_i
		}
	end

	def cmd(cmd, body)
		@cmdbuf << {
			"CmdType" => cmd,
			"Cmd" => body
		}
	end

	def objname()
		@objname
	end

	def key()
		@key
	end

	def data()
		@data
	end

	def results()
		{
			"CB" => @cmdbuf,
			"S" => @state
		}
	end

	def eval(code, key, data)
		@key = key
		@data = data.sort { |x, y| x["Timestamp"] <=> y["Timestamp"] }
		return binding().eval(code)
	end
end

$stdin.binmode
u = MessagePack::Unpacker.new($stdin)
u.each do |obj|
	res = {}
	trigger = obj["Trigger"]
	if not trigger.empty? then
		e = EvalEnv.new obj["S"], obj["Key"]
		obj["IData"].each { |k, v|
			e.eval(trigger, k, v)
		}
		res = e.results()
	end

	res["CB"] = nil if !res["CB"] || res["CB"].empty?
	res["S"] = nil if !res["S"] || res["S"].empty?

	$stdout.binmode
	$stdout.write(MessagePack.pack(res))
	$stdout.flush
end
