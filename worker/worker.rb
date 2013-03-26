#!/usr/bin/env ruby

require 'msgpack'

class EvalEnv
	def initialize(state, host)
		@state = state
		@host = host
		@cmdbuf = []
	end

	def get_state(key)
		elem = @state[key]
		if elem then
			elem['Value']
		else
			nil
		end
	end

	def set_state(key, value)
		@state[key] = {
			'Value' => value,
			'LastUpdate' => Time.now.to_i
		}
	end

	def each_state()
		@state.each { |k, v|
			yield k, v['Value'], ['LastUpdate']
		}
	end

	def cmd(cmd, options = {})
		opts = {}

		options.each { |k, v|
			opts[k] = v.to_s
		}

		@cmdbuf << {
			'Cmd' => cmd,
			'Options' => opts
		}
	end

	def host()
		@host
	end

	def key()
		@key
	end

	def data()
		@data
	end

	def results()
		{
			'CmdBuffer' => @cmdbuf,
			'State' => @state
		}
	end

	def eval(code, key, data)
		@key = key
		@data = data.sort { |x, y| x['Timestamp'] <=> y['Timestamp'] }
		begin
			return binding().eval(code)
		rescue Exception => e
			cmd 'log', :message => "Error: #{e.message}"
		end
	end
end

$stdin.binmode
u = MessagePack::Unpacker.new($stdin)
u.each do |obj|
	res = {}
	trigger = obj['Trigger']
	if not trigger.empty? then
		e = EvalEnv.new obj['State'], obj['Key']
		obj['IData'].each { |k, v|
			e.eval(trigger, k, v)
		}
		res = e.results()
	end

	res['CmdBuffer'] = nil if !res['CmdBuffer'] || res['CmdBuffer'].empty?
	res['State'] = nil if !res['State'] || res['State'].empty?

	$stdout.binmode
	$stdout.write(MessagePack.pack(res))
	$stdout.flush
end
