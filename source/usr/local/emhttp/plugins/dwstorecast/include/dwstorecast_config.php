<?
$dwstorecast_cfg = parse_ini_file("/boot/config/plugins/dwstorecast/dwstorecast.cfg");

$dwstorecast_service = trim(isset($dwstorecast_cfg['DASHBOARD']) ? htmlspecialchars($dwstorecast_cfg['DASHBOARD']) : 'disable');
$dwstorecast_metricsapi = trim(isset($dwstorecast_cfg['METRICSAPI']) ? htmlspecialchars($dwstorecast_cfg['METRICSAPI']) : 'enable');
$dwstorecast_forecastcron = trim(isset($dwstorecast_cfg['FORECASTCRON']) ? htmlspecialchars($dwstorecast_cfg['FORECASTCRON']) : 'daily');
?>
