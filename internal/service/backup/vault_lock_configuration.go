package backup

import (
	"context"
	"log"
	"regexp"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/backup"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

// @SDKResource("aws_backup_vault_lock_configuration")
func ResourceVaultLockConfiguration() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceVaultLockConfigurationCreate,
		ReadWithoutTimeout:   resourceVaultLockConfigurationRead,
		DeleteWithoutTimeout: resourceVaultLockConfigurationDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			"backup_vault_name": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringMatch(regexp.MustCompile(`^[a-zA-Z0-9\-\_\.]{1,50}$`), "must consist of lowercase letters, numbers, and hyphens."),
			},
			"backup_vault_arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"changeable_for_days": {
				Type:         schema.TypeInt,
				Optional:     true,
				ForceNew:     true,
				ValidateFunc: validation.IntAtLeast(3),
			},
			"max_retention_days": {
				Type:     schema.TypeInt,
				Optional: true,
				ForceNew: true,
			},
			"min_retention_days": {
				Type:     schema.TypeInt,
				Optional: true,
				ForceNew: true,
			},
		},
	}
}

func resourceVaultLockConfigurationCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).BackupConn()

	name := d.Get("backup_vault_name").(string)
	input := &backup.PutBackupVaultLockConfigurationInput{
		BackupVaultName: aws.String(name),
	}

	if v, ok := d.GetOk("changeable_for_days"); ok {
		input.ChangeableForDays = aws.Int64(int64(v.(int)))
	}

	if v, ok := d.GetOk("max_retention_days"); ok {
		input.MaxRetentionDays = aws.Int64(int64(v.(int)))
	}

	if v, ok := d.GetOk("min_retention_days"); ok {
		input.MinRetentionDays = aws.Int64(int64(v.(int)))
	}

	_, err := conn.PutBackupVaultLockConfigurationWithContext(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating Backup Vault Lock Configuration (%s): %s", name, err)
	}

	d.SetId(name)

	return append(diags, resourceVaultLockConfigurationRead(ctx, d, meta)...)
}

func resourceVaultLockConfigurationRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).BackupConn()

	output, err := FindVaultByName(ctx, conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] Backup Vault Lock Configuration (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading Backup Vault Lock Configuration (%s): %s", d.Id(), err)
	}

	d.Set("backup_vault_arn", output.BackupVaultArn)
	d.Set("backup_vault_name", output.BackupVaultName)
	d.Set("max_retention_days", output.MaxRetentionDays)
	d.Set("min_retention_days", output.MinRetentionDays)

	return diags
}

func resourceVaultLockConfigurationDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).BackupConn()

	log.Printf("[DEBUG] Deleting Backup Vault Lock Configuration: %s", d.Id())
	_, err := conn.DeleteBackupVaultLockConfigurationWithContext(ctx, &backup.DeleteBackupVaultLockConfigurationInput{
		BackupVaultName: aws.String(d.Id()),
	})

	if tfawserr.ErrCodeEquals(err, backup.ErrCodeResourceNotFoundException) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting Backup Vault Lock Configuration (%s): %s", d.Id(), err)
	}

	return diags
}
