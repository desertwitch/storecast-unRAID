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
$dwstorecast_cfg = parse_ini_file("/boot/config/plugins/dwstorecast/dwstorecast.cfg");

$dwstorecast_dashboard = trim(isset($dwstorecast_cfg['DASHBOARD']) ? htmlspecialchars($dwstorecast_cfg['DASHBOARD']) : 'disable');
$dwstorecast_metricsapi = trim(isset($dwstorecast_cfg['METRICSAPI']) ? htmlspecialchars($dwstorecast_cfg['METRICSAPI']) : 'enable');
$dwstorecast_forecastcron = trim(isset($dwstorecast_cfg['FORECASTCRON']) ? htmlspecialchars($dwstorecast_cfg['FORECASTCRON']) : 'disable');
?>
