# Upstream Simulator

This is the upstream simulator project.  

The upstream simulator is a utility to simulate a bunch of binlog change events for the upstream MySQL servers.  For the motivation of this project, please refer to [here](https://github.com/pingcap/tiflow/issues/4835)

## Features supported now

* Basic DML simulation for multiple tables from multiple data sources.
* Custom workload definition through a domain specific language (DSL).
* Basic DDL manipulation through RESTful API.
* Auto-refresh schema changes for the simulating tables.

## How to build

In the repository root, run ``make dm-simualtor`` to build the source code.  The generated binary will be under `${REPO_ROOT}/bin/dm-simulator`

## How to use it

* Create some empty simuating tables on the upstream MySQL instances.
* Create a configuration file, adding those simulating data sources and tables inside, as well as some workloads.  
  There is an example configuration in the `examples` folder.
* Start the simulator using the configuration file
* After that, some data will be automatically initted on those simulating tables, and then the DML statements are automatically generated according to the workload configuration.
* For schema change, you can use the exposed HTTP API to send some DDL instructions (you can refer to an example script in the `examples` folder),  
  or you can directly change the table structure on the upstream MySQL.  The simulator will automatically detect the latest table structure and adjust the DML statements generated.

### DSL Syntax

The simulator provides a simple domain specific language (DSL) for describing custom workloads.  Here are some further examples:

```
INSERT table_id_01;
UPDATE table_id_01;
DELETE table_id_01;
RANDOM-DML table_id_01;
```

The DSL supports repeated operations.  The repeat operations can be used in a recursive way.
```
REPEAT 3 (
    INSERT table_id_01;
    REPEAT 5 (
        RANDOM-DML table_id_02;
    );
    DELETE table_id_01;
);
```

The DSL supports simple assignments, to assign the specific row for futher operations on the dedicated row.
```
@therow = INSERT table_id_01;
UPDATE @therow;
DELETE @therow;
```

## Current limitations

* The project is in pre-alpha state.
* Only INT and VARCHAR columns are supported for simulation.
* The modification candidate pool (MCP) eviction mechanism is missing.  By design, if the MCP exceeds a capacity limit, the simulator will try to evict some data inside the upstream, so as to limit the disk space usage on the upstream.  So it is not recommend to simulate a INSERT-only workload on the upstream.
* The random property is hard-coded.  The INSERT/UPDATE/DELETE ratio for a RANDOM-DML operation is 1:2:1.  
* The parallel workers for the workload simualations on a data source is hard-coded as 8.

## TODOs

There are some TODOs that can be done in the future.

Short-term:
* Support more data type simulation
* MCP eviction mechanism
* Customized RANDOM-DML attributes
* Configurable parallel workers

Long-term:
* Make it a complete simulation service through API:
    * APIs to start/stop a table's simulation
    * APIs to add/delete workflows on the fly
    * More user-friendly DDL operation interface
    * ...
* Record metadata in a meta-DB.  So that all those configuration can be done dynamically using extra APIs.
* Automatically create tables according to the configurations.
* Automatic upstream provision.  So that if there are several hundres of upstream on a cluster, the users don't need to create all these upstream MySQL instances manually.
* Batch DDL operations on several tables on different data sources, to simulate a batch table structure change on several partitioned tables with the same table structure.
* Rate-limit on the simulation.
* Metrics collection, include statements generated per second, query executed per second, transactions per second, average latency, ...
* For the custom workload definition, replace the simple DSL with a more advanced general-purpose language, like Lua, so that sysbench Lua scripts can be directly used here as the workload.
