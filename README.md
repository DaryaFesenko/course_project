
- Логи, запросы, результаты пишутся в директорию проекта
- Файл csv так же беерется из директории проекта
- Можно указать полный путь для всех файлов или просто их названия в файле config.yaml

Формат запроса, чтобы все сработало:

    SELECT * (или поля через запятую) FROM имя_csv_файла(без .csv) WHERE column_name OP 'example' [AND/OR column_name OP 5];


    - and и or комбинировать нельзя
    - OP: =, <=, =>, <, >, !=(вместо NOT)
    - SELECT, FROM, WHERE, AND, OR можно писать маленькими/большими буквами
    - имена колонок можно писать маленькими/большими буквами
    - работает только со строками и целыми числами
    - в конце строки обязательно ';'


Примеры запросов:
  - Файл covid_19_data
      SELECT * FROM covid_19_data WHERE Country/Region=Canada;
      SELECT ObservationDate, Confirmed FROM covid_19_data WHERE Country/Region != Canada;
      SELECT * FROM covid_19_data WHERE Country/Region != Canada AND Deaths > 30;
    
  - Файл bum24fullexport
      SELECT * FROM business WHERE Magnitude >= 7 OR STATUS = F;
      SELECT status, units FROM business WHERE Suppressed != Y;