<?
/* Copyright Derek Macias (parts of code from NUT package)
 * Copyright macester (parts of code from NUT package)
 * Copyright gfjardim (parts of code from NUT package)
 * Copyright SimonF (parts of code from NUT package)
 * Copyright Dan Landon (parts of code from Web GUI)
 * Copyright Bergware International (parts of code from Web GUI)
 * Copyright Lime Technology (any and all other parts of Unraid)
 *
 * Copyright desertwitch (as author and maintainer of this file)
 *
 * This program is free software; you can redistribute it and/or
 * modify it under the terms of the GNU General Public License 2
 * as published by the Free Software Foundation.
 *
 * The above copyright notice and this permission notice shall be
 * included in all copies or substantial portions of the Software.
 *
 */
    try {
        $storecast_running = !empty(shell_exec("pgrep -x storecast 2>/dev/null"));
        if(file_exists("/tmp/storecast.log")) {
            $modtime_log = filemtime("/tmp/storecast.log");
            if($modtime_log) $modtime_log = date("Y-m-d H:i:s", $modtime_log);
            echo json_encode([
                'running' => $storecast_running,
                'modtime_log' => $modtime_log,
                'response' => htmlspecialchars(file_get_contents("/tmp/storecast.log") ?: "Failed to load logfile.")
            ]);
        } else {
            echo json_encode([
                'running' => $storecast_running,
                'response' => "Nothing here (yet)... - go generate one! :-)"
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
