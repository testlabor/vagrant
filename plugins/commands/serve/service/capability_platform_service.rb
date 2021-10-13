require "google/protobuf/well_known_types"

module VagrantPlugins
  module CommandServe
    module Service
      module CapabilityPlatformService

        def self.included(klass)
          klass.include(Util::ServiceInfo)
          klass.prepend(Util::HasMapper)
          klass.prepend(Util::HasBroker)
          klass.prepend(Util::HasLogger)
          klass.prepend(Util::ExceptionLogger)

          klass.class_eval do
            attr_reader :capabilities, :default_args
          end
        end

        def initialize_capability_platform!(capabilities, default_args)
          @capabilities = capabilities
          @default_args = default_args
        end

        def seed(req, ctx)
          logger.info "seeding this service with values from the client"
          logger.info "values to seed include: #{req.list.inspect}"
          @seeds = req.list.map{ |x| x }
          Empty.new
        end

        def seeds(req, ctx)
          SDK::Args::Direct.new(list: @seeds)
        end

        def has_capability_spec(*_)
          SDK::FuncSpec.new(
            name: "has_capability_spec",
            args: [
              SDK::FuncSpec::Value.new(
                type: "hashicorp.vagrant.sdk.Args.NamedCapability",
                name: "",
              )
            ],
            result: [
              SDK::FuncSpec::Value.new(
                type: "hashicorp.vagrant.sdk.Platform.Capability.CheckResp",
                name: "",
              ),
            ],
          )
        end

        def has_capability(req, ctx)
          with_info(ctx) do |info|
            cap_name = mapper.funcspec_map(req)
            plugin_name = info.plugin_name
            logger.debug("checking for #{cap_name} capability in #{plugin_name}")

            caps_registry = @capabilities[plugin_name.to_sym]
            has_cap = caps_registry.key?(cap_name.to_sym)

            SDK::Platform::Capability::CheckResp.new(
              has_capability: has_cap
            )
          end
        end

        def capability_spec(req, ctx)
          SDK::FuncSpec.new(
            name: "capability_spec",
            args: default_args + [
              SDK::FuncSpec::Value.new(
                type: "hashicorp.vagrant.sdk.Args.Direct",
                name: "",
              )
            ],
            result: [
              SDK::FuncSpec::Value.new(
                type: "hashicorp.vagrant.sdk.Platform.Capability.Resp",
                name: "",
              )
            ]
          )
        end

        def capability(req, ctx)
          with_info(ctx) do |info|
            logger.debug("executing capability, got req #{req}")
            cap_name = req.name.to_sym
            plugin_name = info.plugin_name.to_sym
            caps_registry = capabilities[plugin_name]
            target_cap = caps_registry.get(cap_name)

            args = mapper.funcspec_map(req.func_args, mapper, broker)
            args = [args.first] + args.last
            cap_method = target_cap.method(cap_name)

            result = cap_method.call(*args)

            val = Google::Protobuf::Value.new
            val.from_ruby(result)
            SDK::Platform::Capability::Resp.new(
              result: Google::Protobuf::Any.pack(val)
            )
          end
        end
      end
    end
  end
end
