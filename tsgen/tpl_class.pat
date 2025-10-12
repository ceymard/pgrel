
@/
@func GenerateClass(w io.Writer, rel *db.Relation) {
  export class @rel.Name|snakeCase extends Model {
    @for col in rel.Columns {
      @col.Name|snakeCase
    }
  }
}
