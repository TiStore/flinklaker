package org.apache.hudi;

import org.apache.hudi.factory.CollectSinkTableFactory;

import com.google.gson.Gson;
import org.apache.flink.configuration.Configuration;
import org.apache.flink.core.execution.JobClient;
import org.apache.flink.table.api.EnvironmentSettings;
import org.apache.flink.table.api.TableEnvironment;
import org.apache.flink.table.api.TableResult;
import org.apache.flink.table.api.TableSchema;
import org.apache.flink.table.api.config.ExecutionConfigOptions;
import org.apache.flink.table.api.internal.TableEnvironmentImpl;
import org.apache.flink.table.catalog.ObjectPath;
import org.apache.flink.table.catalog.exceptions.TableNotExistException;
import org.apache.flink.table.types.DataType;
import org.apache.flink.types.Row;
import org.apache.flink.util.CollectionUtil;
import org.springframework.boot.SpringApplication;
import org.springframework.boot.autoconfigure.SpringBootApplication;
import org.springframework.web.bind.annotation.GetMapping;
import org.springframework.web.bind.annotation.RequestParam;
import org.springframework.web.bind.annotation.RestController;

import java.util.Collection;
import java.util.HashMap;
import java.util.List;
import java.util.Map;
import java.util.concurrent.TimeUnit;
import java.util.stream.Collectors;

import static java.lang.Boolean.parseBoolean;

@SpringBootApplication
@RestController
public class HudiApplication {

  public static void main(String[] args) {
    SpringApplication.run(HudiApplication.class, args);
  }

  @GetMapping("/hudi-backend")
  public String hudiBackend(
      @RequestParam(value = "streaming", defaultValue = "true") String streaming)
      throws InterruptedException, TableNotExistException {
    final TableEnvironment tableEnv;
    if (parseBoolean(streaming)) {
      tableEnv = TableEnvironment.create(EnvironmentSettings.newInstance().build());
      tableEnv
          .getConfig()
          .getConfiguration()
          .setInteger(ExecutionConfigOptions.TABLE_EXEC_RESOURCE_DEFAULT_PARALLELISM, 1);
      final Configuration execConf = tableEnv.getConfig().getConfiguration();
      execConf.setString("execution.checkpointing.interval", "2s");
      execConf.setString("restart-strategy", "fixed-delay");
      execConf.setString("restart-strategy.fixed-delay.attempts", "0");
    } else {
      tableEnv =
          TableEnvironmentImpl.create(EnvironmentSettings.newInstance().inBatchMode().build());
      tableEnv
          .getConfig()
          .getConfiguration()
          .setInteger(ExecutionConfigOptions.TABLE_EXEC_RESOURCE_DEFAULT_PARALLELISM, 1);
    }

    final String hoodieTableDDL =
        "CREATE Table hudi_orders(\n"
            + "  order_id INT NOT NULL,,\n"
            + "  car_id INT,\n"
            + "  from_x DOUBLE,\n"
            + "  from_y DOUBLE,\n"
            + "  to_x DOUBLE,\n"
            + "  to_y DOUBLE,\n"
            + "  status STRING,\n"
            + "  create_time TIMESTAMP(3),\n"
            + "  update_time TIMESTAMP(3),\n"
            + "  PRIMARY KEY (`order_id`) NOT ENFORCED\n"
            + ") with (\n"
            + "  'connector' = 'hudi',\n"
            + "  'path' = '///home/ec2-user/data/orders',\n"
            + "  'read.streaming.enabled' = '"
            + streaming
            + "'\n"
            + ")";
    tableEnv.executeSql(hoodieTableDDL);
    final List<Row> rows =
        execSelectSql(
            tableEnv,
            "select * from t1",
            parseBoolean(streaming) ? ExecMode.STREAM : ExecMode.BATCH);
    Map<Integer, Map<String, Float>> trajectories = new HashMap<>();
    for (Row row : rows) {
      Map<String, Float> coordinates = new HashMap<>();
      coordinates.put("from_x", (Float) row.getField(2));
      coordinates.put("from_y", (Float) row.getField(3));
      coordinates.put("to_x", (Float) row.getField(4));
      coordinates.put("to_y", (Float) row.getField(5));
      trajectories.put((Integer) row.getField(0), coordinates);
    }
    return new Gson().toJson(trajectories);
  }

  // -------------------------------------------------------------------------
  //  Utilities
  // -------------------------------------------------------------------------
  private enum ExecMode {
    BATCH,
    STREAM
  }

  private List<Row> execSelectSql(TableEnvironment tEnv, String select, ExecMode execMode)
      throws TableNotExistException, InterruptedException {
    final String[] splits = select.split(" ");
    final String tableName = splits[splits.length - 1];
    switch (execMode) {
      case STREAM:
        return execSelectSql(tEnv, select, 10, tableName);
      case BATCH:
        return CollectionUtil.iterableToList(
            () -> tEnv.sqlQuery("select * from " + tableName).execute().collect());
      default:
        throw new AssertionError();
    }
  }

  private List<Row> execSelectSql(
      TableEnvironment tEnv, String select, long timeout, String sourceTable)
      throws InterruptedException, TableNotExistException {
    final String sinkDDL;
    if (sourceTable != null) {
      ObjectPath objectPath = new ObjectPath(tEnv.getCurrentDatabase(), sourceTable);
      TableSchema schema =
          tEnv.getCatalog(tEnv.getCurrentCatalog()).get().getTable(objectPath).getSchema();
      sinkDDL = getCollectSinkDDL("sink", schema);
    } else {
      sinkDDL = getCollectSinkDDL("sink");
    }
    return execSelectSql(tEnv, select, sinkDDL, timeout);
  }

  private List<Row> execSelectSql(
      TableEnvironment tEnv, String select, String sinkDDL, long timeout)
      throws InterruptedException {
    tEnv.executeSql("DROP TABLE IF EXISTS sink");
    tEnv.executeSql(sinkDDL);
    TableResult tableResult = tEnv.executeSql("insert into sink " + select);
    TimeUnit.SECONDS.sleep(timeout);
    tableResult.getJobClient().ifPresent(JobClient::cancel);
    tEnv.executeSql("DROP TABLE IF EXISTS sink");
    return CollectSinkTableFactory.RESULT.values().stream()
        .flatMap(Collection::stream)
        .collect(Collectors.toList());
  }

  private static String getCollectSinkDDL(String tableName, TableSchema tableSchema) {
    final StringBuilder builder = new StringBuilder("create table " + tableName + "(\n");
    String[] fieldNames = tableSchema.getFieldNames();
    DataType[] fieldTypes = tableSchema.getFieldDataTypes();
    for (int i = 0; i < fieldNames.length; i++) {
      builder.append("  `").append(fieldNames[i]).append("` ").append(fieldTypes[i].toString());
      if (i != fieldNames.length - 1) {
        builder.append(",");
      }
      builder.append("\n");
    }
    final String withProps =
        "" + ") with (\n" + "  'connector' = '" + CollectSinkTableFactory.FACTORY_ID + "'\n" + ")";
    builder.append(withProps);
    return builder.toString();
  }

  private static String getCollectSinkDDL(String tableName) {
    return "create table "
        + tableName
        + "(\n"
        + "  order_id INT NOT NULL,,\n"
        + "  car_id INT,\n"
        + "  from_x DOUBLE,\n"
        + "  from_y DOUBLE,\n"
        + "  to_x DOUBLE,\n"
        + "  to_y DOUBLE,\n"
        + "  status STRING,\n"
        + "  create_time TIMESTAMP(3),\n"
        + "  update_time TIMESTAMP(3),\n"
        + "  PRIMARY KEY (`order_id`) NOT ENFORCED\n"
        + ") with (\n"
        + "  'connector' = '"
        + CollectSinkTableFactory.FACTORY_ID
        + "'"
        + ")";
  }
}
