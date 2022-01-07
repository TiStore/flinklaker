package org.apache.hudi;

import org.apache.hudi.factory.CollectSinkTableFactory;

import com.google.gson.Gson;
import org.apache.flink.configuration.Configuration;
import org.apache.flink.core.execution.JobClient;
import org.apache.flink.table.api.EnvironmentSettings;
import org.apache.flink.table.api.TableEnvironment;
import org.apache.flink.table.api.TableResult;
import org.apache.flink.table.api.config.ExecutionConfigOptions;
import org.apache.flink.table.api.internal.TableEnvironmentImpl;
import org.apache.flink.types.Row;
import org.apache.flink.util.CollectionUtil;
import org.springframework.boot.SpringApplication;
import org.springframework.boot.autoconfigure.SpringBootApplication;
import org.springframework.web.bind.annotation.CrossOrigin;
import org.springframework.web.bind.annotation.GetMapping;
import org.springframework.web.bind.annotation.RequestParam;
import org.springframework.web.bind.annotation.RestController;

import java.util.ArrayList;
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

    @CrossOrigin(origins = "*", maxAge = 3600)
    @GetMapping("/")
    public String backend(
            @RequestParam(value = "streaming", defaultValue = "false") String streaming)
            throws InterruptedException {
        final TableEnvironment tableEnv;
        if (parseBoolean(streaming)) {
            tableEnv = TableEnvironment.create(EnvironmentSettings.newInstance().build());
            tableEnv.getConfig()
                    .getConfiguration()
                    .setInteger(ExecutionConfigOptions.TABLE_EXEC_RESOURCE_DEFAULT_PARALLELISM, 1);
            final Configuration execConf = tableEnv.getConfig().getConfiguration();
            execConf.setString("execution.checkpointing.interval", "2s");
            execConf.setString("restart-strategy", "fixed-delay");
            execConf.setString("restart-strategy.fixed-delay.attempts", "0");
        } else {
            tableEnv =
                    TableEnvironmentImpl.create(
                            EnvironmentSettings.newInstance().inBatchMode().build());
            tableEnv.getConfig()
                    .getConfiguration()
                    .setInteger(ExecutionConfigOptions.TABLE_EXEC_RESOURCE_DEFAULT_PARALLELISM, 1);
        }

        tableEnv.executeSql(
                "CREATE TABLE hudi_orders(\n"
                        + "  order_id INT NOT NULL,\n"
                        + "  car_id INT,\n"
                        + "  from_x DOUBLE,\n"
                        + "  from_y DOUBLE,\n"
                        + "  to_x DOUBLE,\n"
                        + "  to_y DOUBLE,\n"
                        + "  status STRING,\n"
                        + "  create_time TIMESTAMP(3),\n"
                        + "  update_time TIMESTAMP(3),\n"
                        + "  PRIMARY KEY (`order_id`) NOT ENFORCED\n"
                        + ") WITH (\n"
                        + "  'connector' = 'hudi',\n"
                        + "  'path' = '///home/ec2-user/data/orders',\n"
                        + "  'read.streaming.enabled' = '"
                        + streaming
                        + "'\n"
                        + ")");
        tableEnv.executeSql(
                "CREATE TABLE hudi_cars(\n"
                        + "  id INT NOT NULL,\n"
                        + "  location_x DOUBLE,\n"
                        + "  location_y DOUBLE,\n"
                        + "  status STRING,\n"
                        + "  create_time TIMESTAMP(3),\n"
                        + "  update_time TIMESTAMP(3),\n"
                        + "  PRIMARY KEY (`id`) NOT ENFORCED\n"
                        + ") WITH (\n"
                        + "  'connector' = 'hudi',\n"
                        + "  'path' = '///home/ec2-user/data/cars',\n"
                        + "  'read.streaming.enabled' = '"
                        + streaming
                        + "'\n"
                        + ")");
        final List<Row> rows =
                executeSelectSql(
                        tableEnv, parseBoolean(streaming) ? ExecMode.STREAM : ExecMode.BATCH);
        List<Map<String, Map<String, Map<String, String>>>> cars = new ArrayList<>();
        for (Row row : rows) {
            Map<String, Map<String, Map<String, String>>> car = new HashMap<>();
            Map<String, Map<String, String>> info = new HashMap<>();
            Map<String, String> detail = new HashMap<>();
            detail.put("coordinates", String.format("%s,%s", row.getField(2), row.getField(1)));
            detail.put("status", String.valueOf(row.getField(3)));
            info.put("detail", detail);
            if (row.getField(4) != null) {
                Map<String, String> order = new HashMap<>();
                order.put("id", String.valueOf(row.getField(4)));
                order.put(
                        "coordinates",
                        String.format(
                                "%s,%s|%s,%s",
                                row.getField(6),
                                row.getField(5),
                                row.getField(8),
                                row.getField(7)));
                info.put("order", order);
            }
            car.put(String.valueOf(row.getField(0)), info);
            cars.add(car);
        }
        System.out.println(new Gson().toJson(cars));
        return new Gson().toJson(cars);
    }

    // -------------------------------------------------------------------------
    //  Utilities
    // -------------------------------------------------------------------------
    private enum ExecMode {
        BATCH,
        STREAM
    }

    private List<Row> executeSelectSql(TableEnvironment tEnv, ExecMode execMode)
            throws InterruptedException {
        final String selectSql =
                "SELECT c.id, c.location_x, c.location_y, c.status, o.order_id, o.from_x, o.from_y, o.to_x,  o.to_y, o.status "
                        + "FROM hudi_cars as c "
                        + "LEFT JOIN hudi_orders as o "
                        + "ON c.id=o.car_id AND o.status='running'";
        switch (execMode) {
            case STREAM:
                tEnv.executeSql("DROP TABLE IF EXISTS sink");
                tEnv.executeSql(getCollectSinkDDL());
                TableResult tableResult = tEnv.executeSql("INSERT INTO sink " + selectSql);
                TimeUnit.SECONDS.sleep(10);
                tableResult.getJobClient().ifPresent(JobClient::cancel);
                return CollectSinkTableFactory.RESULT.values().stream()
                        .flatMap(Collection::stream)
                        .collect(Collectors.toList());
            case BATCH:
                return CollectionUtil.iterableToList(
                        () -> tEnv.sqlQuery(selectSql).execute().collect());
            default:
                throw new AssertionError();
        }
    }

    private String getCollectSinkDDL() {
        return "CREATE TABLE sink(\n"
                + "  id INT NOT NULL,\n"
                + "  location_x DOUBLE,\n"
                + "  location_y DOUBLE,\n"
                + "  car_status STRING,\n"
                + "  order_id INT,\n"
                + "  from_x DOUBLE,\n"
                + "  from_y DOUBLE,\n"
                + "  to_x DOUBLE,\n"
                + "  to_y DOUBLE\n"
                + ") WITH (\n"
                + "  'connector' = '"
                + CollectSinkTableFactory.IDENTIFIER
                + "'"
                + ")";
    }
}
