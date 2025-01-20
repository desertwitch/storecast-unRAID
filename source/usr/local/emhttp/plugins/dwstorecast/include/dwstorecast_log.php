<?php
    try {
        $storecast_running = !empty(shell_exec("pgrep -x storecast 2>/dev/null"));
        if(file_exists("/tmp/storecast.log")) {
            $modtime_log = filemtime("/tmp/storecast.log");
            if($modtime_log) $modtime_log = date("Y-m-d H:i:s", $modtime_log);

            $modtime_json = false;
            if(file_exists("/tmp/storecast.json")) {
                $modtime_json = filemtime("/tmp/storecast.json");
                if($modtime_json) $modtime_json = date("Y-m-d H:i:s", $modtime_json);
            }
            
            echo json_encode([
                'running' => $storecast_running,
                'modtime_log' => $modtime_log,
                'modtime_json' => $modtime_json,
                'response' => htmlspecialchars(file_get_contents("/tmp/storecast.log") ?: "Failed to load logfile.")
            ]);
        } else {
            $modtime_json = false;
            if(file_exists("/tmp/storecast.json")) {
                $modtime_json = filemtime("/tmp/storecast.json");
                if($modtime_json) $modtime_json = date("Y-m-d H:i:s", $modtime_json);
            }
            echo json_encode([
                'running' => $storecast_running,
                'modtime_json' => $modtime_json,
                'response' => "Waiting for Forecast..."
            ]);
        }
    } catch(\Throwable $t) {
        error_log($t);
        echo json_encode([
            'running' => $storecast_running,
            'response' => htmlspecialchars($t->getMessage())
        ]);
    } catch(\Exception $e) {
        error_log($e);
        echo json_encode([
            'running' => $storecast_running,
            'response' => htmlspecialchars($e->getMessage())
        ]);
    }
?>
