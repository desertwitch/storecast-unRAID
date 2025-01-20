<?php
    try {
        $storecast_running = !empty(shell_exec("pgrep -x storecast 2>/dev/null"));
        if(file_exists("/tmp/storecast.log")) {
            echo json_encode([
                'running' => $storecast_running,
                'response' => htmlspecialchars(file_get_contents("/tmp/storecast.log") ?: "Failed to load logfile.")
            ]);
        } else {
            echo json_encode([
                'running' => $storecast_running,
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
